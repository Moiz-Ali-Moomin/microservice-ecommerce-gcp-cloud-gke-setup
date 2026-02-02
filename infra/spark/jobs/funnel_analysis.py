# infra/spark/jobs/funnel_analysis.py
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, window, count

def main():
    spark = SparkSession.builder \
        .appName("FunnelAnalysis") \
        .getOrCreate()

    # Read events from GCS (Partitioned by date/hour)
    events_df = spark.read.json("gs://ecommerce-events-archive/raw/*/*/*/*.json")

    # Funnel Calculation: Page View -> CTA Click -> Conversion
    funnel_stats = events_df.groupBy("campaign_id", "event_type") \
        .agg(count("event_id").alias("count")) \
        .orderBy("campaign_id")

    # Write results to Postgres (via JDBC)
    funnel_stats.write \
        .format("jdbc") \
        .option("url", "jdbc:postgresql://postgres-primary.ecommerce.svc.cluster.local:5432/ecommerce_db") \
        .option("dbtable", "analytics.funnel_daily") \
        .option("user", "app_user") \
        .option("password", "secret") \
        .mode("append") \
        .save()

    spark.stop()

if __name__ == "__main__":
    main()
