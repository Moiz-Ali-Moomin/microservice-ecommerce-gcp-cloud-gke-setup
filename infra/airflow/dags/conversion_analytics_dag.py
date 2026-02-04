# infra/airflow/dags/conversion_analytics_dag.py
from datetime import timedelta
from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.spark_kubernetes import SparkKubernetesOperator
from airflow.providers.cncf.kubernetes.sensors.spark_kubernetes import SparkKubernetesSensor
from airflow.utils.dates import days_ago

default_args = {
    'owner': 'data-eng',
    'depends_on_past': False,
    'start_date': days_ago(1),
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
    'execution_timeout': timedelta(hours=2), # Prevent stuck jobs
}

with DAG(
    'conversion_funnel_daily',
    default_args=default_args,
    schedule_interval='@daily',
    description='Aggregates Kafka events from GCS to calculate funnel conversion rates',
    catchup=False,
    tags=['spark', 'analytics', 'funnel'],
) as dag:

    # 1. Submit the SparkApplication to K8s
    submit_job = SparkKubernetesOperator(
        task_id='submit_funnel_analysis',
        namespace='analytics',
        application_file='spark-submit.yaml', # GitSync ensures this file is present
        do_xcom_push=True,
    )

    # 2. Wait for completion (Reschedule mode saves worker slot while waiting)
    monitor_job = SparkKubernetesSensor(
        task_id='monitor_funnel_analysis',
        namespace='analytics',
        application_name="{{ task_instance.xcom_pull(task_ids='submit_funnel_analysis')['metadata']['name'] }}",
        attach_log=True,
        mode='reschedule', # CRITICAL: Releases worker slot during wait
        poke_interval=60,
        timeout=7200, # 2 hours
    )

    submit_job >> monitor_job
