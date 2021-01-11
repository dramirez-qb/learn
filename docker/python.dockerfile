FROM python:3.9-slim

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

RUN mkdir -p /app
WORKDIR /app

COPY requirements.txt /app/requirements.txt
COPY app.py /app/app.py
COPY gunicorn_config.py /app/gunicorn_config.py

RUN pip install -r /app/requirements.txt && \
    chmod 777 -R /app/*

ENV LOGLEVEL WARNING

EXPOSE 8000

CMD [ "gunicorn", "-c", "gunicorn_config.py", "app:app" ]
