FROM python:3.6-slim

WORKDIR /workspace

# Copy the requirements.txt
COPY requirements.txt requirements.txt

RUN pip install -r requirements.txt

# Copy reset of the code
COPY . .

CMD python main.py