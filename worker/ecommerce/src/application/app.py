import os
from celery import Celery

celery = Celery("ecommerce", broker=os.environ["CELERY_BROKER_URL"])
celery.conf.imports = ("application.consumer",)

