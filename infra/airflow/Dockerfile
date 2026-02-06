FROM apache/airflow:2.9.0

# Install the missing Spark provider
# user is 'airflow' by default in the base image
RUN pip install --no-cache-dir apache-airflow-providers-apache-spark
