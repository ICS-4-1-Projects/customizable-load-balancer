# testing.py
from .config import Config
import os

class TestingConfig(Config):
    TESTING = False
    DEBUG = False
    SQLALCHEMY_DATABASE_URI = os.environ.get('TEST_DATABASE_URL', 'sqlite:///:memory:')
    WTF_CSRF_ENABLED = False  # Disables CSRF tokens in the forms for testing purposes
