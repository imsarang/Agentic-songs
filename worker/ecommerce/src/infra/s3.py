from infra.ports import RawStore

import os
import boto3

class ObjectStore(RawStore):
    def  __init__(self) -> None:
        self._client = None
    
    def client(self):
        if self._client is None:
            self._client = boto3.client(
               "s3",
                endpoint_url=os.environ["S3_ENDPOINT"],
                aws_access_key_id=os.environ["S3_ACCESS_KEY"],
                aws_secret_access_key=os.environ["S3_SECRET_KEY"],
                region_name=os.environ.get("S3_REGION", "us-east-1"),
            )
        return self._client

    def get_object(self, key: str) ->  bytes:
        body = self.client().get_object(
            Bucket=os.environ.get("S3_BUCKET", "shanks-raw"),
            Key=key,
        )["Body"]
        return body.read()
    