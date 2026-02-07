from datetime import datetime, timedelta
import json

from airflow import DAG
from airflow.providers.http.operators.http import SimpleHttpOperator

default_args = {
    "owner": "data-eng",
    "depends_on_past": False,
    "start_date": datetime(2024, 1, 1),
    "retries": 1,
    "retry_delay": timedelta(minutes=1),
}

SPARK_MASTER_URL = "http://spark-master:6066"
JOB_FILE = "local:///opt/spark/jobs/funnel_analysis.py"
JOB_NAME = "funnel-analysis-local"

with DAG(
    dag_id="dev_conversion_funnel_daily",
    default_args=default_args,
    schedule="@daily",
    catchup=False,
    description="Local Dev DAG: Submit Spark job via Spark Standalone REST API",
    tags=["spark", "local"],
) as dag:

    submit_job_payload = {
        "action": "CreateSubmissionRequest",
        "appResource": JOB_FILE,
        "appArgs": [],
        "clientSparkVersion": "3.5.1",
        "mainClass": "org.apache.spark.deploy.SparkSubmit",
        "environmentVariables": {
            "SPARK_ENV_LOADED": "1",
            "EVENTS_INPUT_PATH": "/opt/spark/data/raw/events/*.json",
        },
        "sparkProperties": {
            "spark.master": "spark://spark-master:7077",
            "spark.app.name": JOB_NAME,
            "spark.submit.deployMode": "cluster",

            # Executor env vars
            "spark.executorEnv.DB_URL": "jdbc:postgresql://postgres:5432/ecommerce",
            "spark.executorEnv.DB_USER": "user",
            "spark.executorEnv.DB_PASSWORD": "password",
            "spark.executorEnv.EVENTS_INPUT_PATH": "/opt/spark/data/raw/events/*.json",
        },
    }

    submit_job = SimpleHttpOperator(
        task_id="submit_spark_job",
        http_conn_id="spark_master_conn",
        endpoint="/v1/submissions/create",
        method="POST",
        data=json.dumps(submit_job_payload),
        headers={"Content-Type": "application/json"},
        log_response=True,
        response_check=lambda r: r.json().get("success") is True,
    )
