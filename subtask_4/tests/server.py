from flask import Flask
import logging

def create_app(name, port):
    app = Flask(__name__)
    
    @app.route('/')
    def hello():
        return f'{name}'
    
    # Disable logging
    log = logging.getLogger('werkzeug')
    log.disabled = True
    app.logger.disabled = True
    
    app.run(port=port)
