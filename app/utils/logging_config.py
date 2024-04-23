import os
import logging.config
from pythonjsonlogger import jsonlogger

class CustomJsonFormatter(jsonlogger.JsonFormatter):
    def __init__(self, fmt="%(asctime)s %(levelname)s %(name)s %(message)s", style='%', **kwargs):
        super().__init__(fmt=fmt, style=style, **kwargs)

    def add_fields(self, log_record, record, message_dict):
        super(CustomJsonFormatter, self).add_fields(log_record, record, message_dict)
        app_name = os.getenv('APP_NAME', 'GWDynamics')
        log_record['app_name'] = app_name


def setup_logging():
    logs_directory = 'logs'
    app_name = os.getenv('APP_NAME', 'GWDynamics')
    if not os.path.exists(logs_directory):
        os.mkdir(logs_directory)

    logging_configuration = {
        'version': 1,
        'disable_existing_loggers': False,
        'formatters': {
            'json': {
                '()': CustomJsonFormatter,
                'format': '%(asctime)s %(levelname)s %(name)s %(filename)s:%(lineno)d %(message)s',
                'datefmt': '%Y-%m-%d %H:%M:%S',
            },
        },
        'handlers': {
            'file_handler': {
                'class': 'logging.handlers.RotatingFileHandler',
                'level': 'INFO',
                'formatter': 'json',
                'filename': os.path.join(logs_directory, f'{app_name}.log'),
                'maxBytes': 10485760,  # 10MB
                'backupCount': 10,
            },
            'error_file_handler': {
                'class': 'logging.handlers.RotatingFileHandler',
                'level': 'ERROR',
                'formatter': 'json',
                'filename': os.path.join(logs_directory, f'{app_name}-error.log'),
                'maxBytes': 10485760,  # 10MB
                'backupCount': 10,
            },
        },
        'loggers': {
            '': {  # root logger
                'handlers': ['file_handler', 'error_file_handler'],
                'level': 'INFO',
                'propagate': False,
            },
        },
    }

    logging.config.dictConfig(logging_configuration)
