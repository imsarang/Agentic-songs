from celery import Celery
import os

celery = Celery("crawl", broker = os.environ["CELERY_BROKER_URL"])
celery.conf.imports = ("application.consumer",)
