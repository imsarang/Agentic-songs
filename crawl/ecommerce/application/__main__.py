from application.app import celery

def main() -> None:
    celery.worker_main(
        argv=["worker", "--loglevel=INFO", "-Q", "crawl", "--pool=solo"]
    )

if __name__ == "__main__":
    main()