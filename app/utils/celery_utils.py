from celery import Celery

def make_celery(app):
    """Create and configure a Celery object from a Flask application instance."""
    
    # Initialize the Celery object with the application's import name, broker URL, and result backend URL
    # These are supposed to be defined in the app's configuration
    celery = Celery(app.import_name,
                    broker=app.config['CELERY_BROKER_URL'],
                    backend=app.config['CELERY_RESULT_BACKEND'])
    
    # Update Celery's configuration from the Flask application's configuration
    celery.conf.update(app.config)

    # Subclass the Task class to attach the Flask app context to Celery tasks
    class ContextTask(celery.Task):
        """Context task type that wraps the task execution in a Flask application context."""
        def __call__(self, *args, **kwargs):
            with app.app_context():
                return self.run(*args, **kwargs)

    # Set the custom task class as the base class for all Celery tasks
    celery.Task = ContextTask

    return celery
