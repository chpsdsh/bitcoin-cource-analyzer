FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

RUN addgroup --system app \
    && adduser --system --ingroup app app \
    && mkdir -p /data \
    && chown -R app:app /data

COPY app.py .

USER app

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8080"]
