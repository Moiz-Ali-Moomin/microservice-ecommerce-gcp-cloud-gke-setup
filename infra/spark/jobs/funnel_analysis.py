# infra/spark/jobs/funnel_analysis.py
import os
import sys
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, window, count

def main():
    # 1. Initialize Spark
    spark = SparkSession.builder \
        .appName("FunnelAnalysis") \
        .getOrCreate()
    
    # 2. Secure Credential Access (From Env Vars)
    db_user = os.environ.get("DB_USER")
    db_pass = os.environ.get("DB_PASSWORD")
    db_url = os.environ.get("DB_URL", "jdbc:postgresql://postgres:5432/ecommerce_db")
    input_path = os.environ.get("EVENTS_INPUT_PATH", "gs://ecommerce-events-archive/raw/*/*/*/*.json")

    if not db_user or not db_pass:
        print("Error: DB credentials not found in environment variables.")
        sys.exit(1)

    try:
        # 3. Read events from Kafka
        kafka_brokers = os.environ.get("KAFKA_BROKERS", "kafka:9092")
        print(f"Reading events from Kafka: {kafka_brokers}")
        
        # Subscribe to all relevant topics
        topics = "page.viewed,conversion.completed,checkout.started"
        
        # Read from Kafka (Batch Mode - reading current offsets)
        # Note: In a real batch pipeline, we'd manage starting/ending offsets manually or use trigger(once=True) with readStream.
        # For simplicity here (as requested "Spark (batch processing)"), we read the current state of the topic as a batch dataframe.
        raw_df = spark.read \
            .format("kafka") \
            .option("kafka.bootstrap.servers", kafka_brokers) \
            .option("subscribe", topics) \
            .option("startingOffsets", "earliest") \
            .option("endingOffsets", "latest") \
            .load()
            
        # Deserialize JSON value
        from pyspark.sql.functions import from_json, col
        from pyspark.sql.types import StructType, StringType, MapType

        # Schema matches shared-lib Event struct
        schema = StructType() \
            .add("event_id", StringType()) \
            .add("timestamp", StringType()) \
            .add("service", StringType()) \
            .add("event_type", StringType()) \
            .add("session_id", StringType()) \
            .add("offer_id", StringType()) \
            .add("metadata", MapType(StringType(), StringType()))

        events_df = raw_df.select(
            col("topic").alias("event_type"), # Use Kafka Topic as the authoritative Event Type
            from_json(col("value").cast("string"), schema).alias("data")
        ).select(
            "event_type",
            col("data.event_id"), 
            col("data.offer_id"),
            col("data.session_id")
        )

        # 4. Funnel Calculation
        # 4. Funnel Calculation
        # Group by offer_id (campaign) and event_type
        funnel_stats = events_df.groupBy("offer_id", "event_type") \
            .agg(count("event_id").alias("count")) \
            .orderBy("offer_id") \
            .withColumnRenamed("offer_id", "campaign_id") # Map offer_id to campaign_id for DB schema compatibility

        # 5. Write results to Postgres
        # Mode: Overwrite ensures idempotency for re-runs of the same dataset.
        # Ideally, we would use dbtable="analytics.funnel_daily_temp" and swap, 
        # but overwrite is acceptable for this aggregation pattern.
        funnel_stats.write \
            .format("jdbc") \
            .option("url", db_url) \
            .option("dbtable", "analytics.funnel_daily") \
            .option("user", db_user) \
            .option("password", db_pass) \
            .option("driver", "org.postgresql.Driver") \
            .mode("overwrite") \
            .save()
            
        print("Funnel analysis completed successfully.")

    except Exception as e:
        print(f"Error during funnel analysis: {str(e)}")
        sys.exit(1)
        
    finally:
        spark.stop()

if __name__ == "__main__":
    main()
