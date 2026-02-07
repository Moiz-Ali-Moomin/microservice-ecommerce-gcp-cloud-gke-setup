# infra/airflow/dags/conversion_analytics_dag.py

from datetime import timedelta
import os

from airflow import DAG
from airflow.utils.dates import days_ago

# Kubernetes Spark Operator (used conditionally)
from airflow.providers.cncf.kubernetes.operators.spark_kubernetes import (
    SparkKubernetesOperator,
)
from airflow.providers.cncf.kubernetes.sensors.spark_kubernetes import (
    SparkKubernetesSensor,
)

default_args = {
    "owner": "data-eng",
    "depends_on_past": False,
    "start_date": days_ago(1),
    "retries": 3,
    "retry_delay": timedelta(minutes=5),
    "execution_timeout": timedelta(hours=2),  # Prevent stuck jobs
}

with DAG(
    dag_id="conversion_funnel_daily",
    default_args=default_args,
    schedule_interval="@daily",
    description="Aggregates Kafka events from GCS to calculate funnel conversion rates",
    catchup=False,
    tags=["spark", "analytics", "funnel"],
) as dag:

    airflow_env = os.environ.get("AIRFLOW_ENV", "vm")  # default = VM for local dev

    if airflow_env == "k8s":
        # -------------------------------
        # Kubernetes Mode (Spark Operator)
        # -------------------------------
        submit_job = SparkKubernetesOperator(
            task_id="submit_funnel_analysis",
            namespace="analytics",
            application_file="spark-submit.yaml",  # Provided via GitSync
            do_xcom_push=True,
        )

        monitor_job = SparkKubernetesSensor(
            task_id="monitor_funnel_analysis",
            namespace="analytics",
            application_name="{{ task_instance.xcom_pull(task_ids='submit_funnel_analysis')['metadata']['name'] }}",
            attach_log=True,
            mode="reschedule",
            poke_interval=60,
            timeout=7200,
        )

        submit_job >> monitor_job

    else:
        # --------------------------------
        # VM / Docker Compose Mode (SparkSubmit)
        # --------------------------------
        from airflow.providers.apache.spark.operators.spark_submit import (
            SparkSubmitOperator,
        )

        submit_job = SparkSubmitOperator(
            task_id="submit_funnel_analysis",
            application="/opt/spark/jobs/funnel_analysis.py",  # Mounted volume
            conn_id="spark_default",
            conf={
                "spark.master": "spark://spark-master:7077",
                "spark.submit.deployMode": "client",
                "spark.driver.host": "airflow-worker",
            },
            packages=(
                "org.apache.spark:spark-sql-kafka-0-10_2.12:3.5.1,"
                "org.postgresql:postgresql:42.6.0"
            ),
            application_args=[],
        )
