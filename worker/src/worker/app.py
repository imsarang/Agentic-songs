import os

from celery import Celery

celery = Celery("worker", broker=os.environ["CELERY_BROKER_URL"])
celery.conf.imports = ("worker.tasks",)
