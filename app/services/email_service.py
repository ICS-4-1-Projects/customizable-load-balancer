from flask_mail import Message
from flask import current_app

def send_async_email(email_data):
    from app import create_app  # Import here to avoid circular dependency
    app = create_app()  # Only if needed to create a context, otherwise use the current_app context
    celery = app.celery  # Access the Celery instance attached to the app

    @celery.task
    def deliver_email():
        with app.app_context() or current_app.app_context():
            mail = current_app.mail
            msg = Message(
                subject=email_data['subject'],
                sender=email_data['sender'],
                recipients=[email_data['recipient']]
            )
            msg.body = email_data['body']
            mail.send(msg)

    deliver_email.delay()  # Use the delay method to send the email asynchronously
