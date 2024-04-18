from flask import Flask
import os
from .utils.logging_config import setup_logging
from .config import development, testing, production
from .extensions import db, migrate, csrf
from flask_mail import Mail


# Initialize global logging
setup_logging()

def create_app():
    app = Flask(__name__)

    # Load environment configuration
    flask_env = os.getenv('FLASK_ENV', 'production').lower()

    config_mapping = {
        'development': development.DevelopmentConfig,
        'testing': testing.TestingConfig,
        'production': production.ProductionConfig
    }
    app.config.from_object(config_mapping.get(flask_env, 'config.production.ProductionConfig'))
        

    # Initialize FlaskSQLAlchemy and FlaskMigrate
    db.init_app(app)
    from . import models
    migrate.init_app(app, db)

    # Enabling global CSRF Protection
    csrf.init_app(app)
    
    
    # Initialize Flask-Mail
    app.mail = Mail(app)

    
    # Register Blueprints
    from .products import products
    from .errors.handlers import error_bp

    app.register_blueprint(error_bp)
    app.register_blueprint(products, url_prefix='/products')
    
    # Initialize Celery after all other initializations
    from .utils.celery_utils import make_celery
    app.celery = make_celery(app)

    return app
