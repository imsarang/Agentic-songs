import os
import boto3

def get_object(key: str) -> bytes:
    client = boto3.client(
        "s3",
        endpoint_url=os.environ["S3_ENDPOINT"],
        aws_access_key_id=os.environ["S3_ACCESS_KEY"],
        aws_secret_access_key=os.environ["S3_SECRET_KEY"],
        region_name=os.environ.get("S3_REGION", "us-east-1"),
    )
    body = client.get_object(Bucket=os.environ.get("S3_BUCKET", "shanks-raw"), Key=key)["Body"]
    return body.read()