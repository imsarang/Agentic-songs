import boto3
from botocore.exceptions import ClientError

from crawl.config import Settings


class S3Store:
    def __init__(self, settings: Settings) -> None:
        self.bucket = settings.s3_bucket
        self.client = boto3.client(
            "s3",
            endpoint_url=settings.s3_endpoint,
            aws_access_key_id=settings.s3_access_key,
            aws_secret_access_key=settings.s3_secret_key,
            region_name=settings.s3_region,
        )

    def ensure_bucket(self) -> None:
        try:
            self.client.head_bucket(Bucket=self.bucket)
        except ClientError:
            self.client.create_bucket(Bucket=self.bucket)

    def put(self, key: str, body: bytes, content_type: str, meta: dict[str, str]) -> None:
        self.client.put_object(
            Bucket=self.bucket,
            Key=key,
            Body=body,
            ContentType=content_type,
            Metadata=meta,
        )

    def raw_for_url(self, source_id: str, url: str) -> str | None:
        listed = self.client.list_objects_v2(
            Bucket = self.bucket,
            Prefix = f"raw/{source_id}/"
        )

        for item in listed.get("Contents", []):
            head = self.client.head_object(
                Bucket = self.bucket,
                Key = item["Key"]
            )
            if head["Metadata"].get("url") == url:
                return item["Key"]
        return None