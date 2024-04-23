# production.py
from .config import Config
import os

class ProductionConfig(Config):
    DEBUG = False
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL', Config.SQLALCHEMY_DATABASE_URI)
    SESSION_COOKIE_SECURE = True
    REMEMBER_COOKIE_SECURE = True
