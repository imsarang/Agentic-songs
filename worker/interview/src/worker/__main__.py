from worker.app import celery

def main() -> None:
    celery.worker_main(argv=["worker", "--loglevel=INFO", "-Q", "celery", "--pool=solo"])

if __name__ == "__main__":
    main()