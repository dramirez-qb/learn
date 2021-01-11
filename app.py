from flask import Flask, jsonify
app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello World!"

@app.route("/ping")
def ping():
    return "pong"

@app.route("/healthz")
def healthz():
    return jsonify({"alive": True})

if __name__ == '__main__':
    app.run(host='0.0.0.0', debug=True)