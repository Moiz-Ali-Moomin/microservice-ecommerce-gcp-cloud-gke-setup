# infra/airflow/dags/dev_conversion_analytics_dag.py
from datetime import datetime, timedelta
import json
from airflow import DAG
from airflow.providers.http.operators.http import SimpleHttpOperator
from airflow.sensors.http_sensor import HttpSensor

default_args = {
    'owner': 'data-eng',
    'depends_on_past': False,
    'start_date': datetime(2024, 1, 1),
    'retries': 1,
    'retry_delay': timedelta(minutes=1),
}

# Configuration for Local Spark Submission
SPARK_MASTER_URL = "http://spark-master:6066"
JOB_FILE = "local:///opt/spark/jobs/funnel_analysis.py"
JOB_NAME = "funnel-analysis-local"

with DAG(
    'dev_conversion_funnel_daily',
    default_args=default_args,
    schedule_interval='@daily',
    description='Local Dev DAG: Aggregates events via Spark Master REST API',
    catchup=False,
    tags=['spark', 'local'],
) as dag:

    # 1. Submit Job via Spark Master REST API
    # Docs: https://spark.apache.org/docs/latest/spark-standalone.html#rest-api
    submit_job_payload = json.dumps({
        "action": "CreateSubmissionRequest",
        "appArgs": [],
        "appResource": JOB_FILE,
        "clientSparkVersion": "3.5.1",
        "environmentVariables": {
            "SPARK_ENV_LOADED": "1",
            "EVENTS_INPUT_PATH": "/opt/spark/data/raw/events/*.json" # Matches mounted path
        },
        "mainClass": "org.apache.spark.deploy.SparkSubmit", # Not used for Python but required
        "sparkProperties": {
            "spark.master": "spark://spark-master:7077",
            "spark.app.name": JOB_NAME,
            "spark.submit.deployMode": "cluster",
            "spark.driver.supervise": "false",
            # DB Connection Config
            "spark.executorEnv.DB_URL": "jdbc:postgresql://postgres:5432/ecommerce",
            "spark.executorEnv.DB_USER": "user",
            "spark.executorEnv.DB_PASSWORD": "password",
            "spark.executorEnv.EVENTS_INPUT_PATH": "/opt/spark/data/raw/events/*.json"
        }
    })

    submit_job = SimpleHttpOperator(
        task_id='submit_spark_job',
        http_conn_id='spark_master_conn', # We will need to assume or create this connection, or use direct URL
        endpoint='/v1/submissions/create',
        method='POST',
        data=submit_job_payload,
        headers={"Content-Type": "application/json;charset=UTF-8"},
        response_check=lambda response: response.json().get('success') == True,
        log_response=True,
    )
    
    # Note: In a real setup, we would add a Sensor here to poll /v1/submissions/status/{submissionId}
    # For this simple tasks, we just fire and forget, assuming the user checks the Spark UI.

