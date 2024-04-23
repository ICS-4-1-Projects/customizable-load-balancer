# app/errors/handlers.py

from flask import render_template, Blueprint, current_app, request

error_bp = Blueprint('errors', __name__)

@error_bp.app_errorhandler(404)
def not_found_error(error):
    # current_app.logger.error(f'404 Not Found: {request.path}, Error: {error}')
    return '404 Page not found'

@error_bp.app_errorhandler(500)
def internal_error(error):
    current_app.logger.error(f'500 Internal Server Error: {request.path}, Error: {error}', exc_info=True)
    return render_template('errors/500.html'), 500

@error_bp.app_errorhandler(403)
def forbidden_error(error):
    current_app.logger.error(f'403 Forbidden: {request.path}, Error: {error}')
    return render_template('errors/403.html'), 403

@error_bp.app_errorhandler(400)
def bad_request(error):
    current_app.logger.error(f'400 Bad Request: {request.path}, Error: {error}')
    return render_template('errors/400.html'), 400
