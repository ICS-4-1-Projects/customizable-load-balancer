# development.py
from .config import Config
import os

class DevelopmentConfig(Config):
    DEBUG = True
    
    SESSION_COOKIE_SECURE = False
    REMEMBER_COOKIE_SECURE = False
    SESSION_COOKIE_HTTPONLY = False
    REMEMBER_COOKIE_HTTPONLY = False
    
    
    SQLALCHEMY_DATABASE_URI = os.environ.get('DEV_DATABASE_URL', Config.SQLALCHEMY_DATABASE_URI)
    SQLALCHEMY_ECHO = True
    
    
