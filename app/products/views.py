from flask import render_template
from . import products

# Loads the products list page
@products.route('/')
def index():
    return render_template('products/products.html', title='Products')
