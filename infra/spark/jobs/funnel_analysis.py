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
    db_url = os.environ.get("DB_URL", "jdbc:postgresql://postgres-primary.data-postgres.svc.cluster.local:5432/ecommerce_db")
    input_path = os.environ.get("EVENTS_INPUT_PATH", "gs://ecommerce-events-archive/raw/*/*/*/*.json")

    if not db_user or not db_pass:
        print("Error: DB credentials not found in environment variables.")
        sys.exit(1)

    try:
        # 3. Read events
        print(f"Reading events from: {input_path}")
        events_df = spark.read.json(input_path)

        # 4. Funnel Calculation
        funnel_stats = events_df.groupBy("campaign_id", "event_type") \
            .agg(count("event_id").alias("count")) \
            .orderBy("campaign_id")

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
