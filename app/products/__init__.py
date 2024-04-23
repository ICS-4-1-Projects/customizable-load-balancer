from flask import Blueprint

products = Blueprint('products', __name__, template_folder='templates')

from . import views
