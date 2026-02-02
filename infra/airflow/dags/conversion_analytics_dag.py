# infra/airflow/dags/conversion_analytics_dag.py
from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.spark_kubernetes import SparkKubernetesOperator
from airflow.providers.cncf.kubernetes.sensors.spark_kubernetes import SparkKubernetesSensor
from airflow.utils.dates import days_ago

default_args = {
    'owner': 'data-eng',
    'depends_on_past': False,
    'start_date': days_ago(1),
    'retries': 3,
}

with DAG(
    'conversion_funnel_daily',
    default_args=default_args,
    schedule_interval='@daily',
    description='Aggregates Kafka events from GCS to calculate funnel conversion rates',
    catchup=False,
) as dag:

    submit_job = SparkKubernetesOperator(
        task_id='submit_funnel_analysis',
        namespace='analytics',
        application_file='spark-submit.yaml', # Maps to the file in the repo
        do_xcom_push=True,
    )

    monitor_job = SparkKubernetesSensor(
        task_id='monitor_funnel_analysis',
        namespace='analytics',
        application_name="{{ task_instance.xcom_pull(task_ids='submit_funnel_analysis')['metadata']['name'] }}",
        attach_log=True,
    )

    submit_job >> monitor_job
