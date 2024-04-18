# Use an official Python runtime as a base image, specifically a slim version for smaller image size.
FROM python:3.10-slim

# Set environment variables to ensure Python runs in unbuffered mode, recommended for Docker.
# This also prevents Python from writing .pyc files which are unnecessary in this context.
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

# Create and set the working directory
WORKDIR /usr/src/app

# Install system dependencies
# Ensure that the packages installed are minimal and clean up the cache to reduce the layer size.
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY ./requirements.txt /usr/src/app/requirements.txt
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir -r requirements.txt


# Copy the rest of the application into the container
COPY . /usr/src/app/

# Create a non-root user and switch to it for security reasons
RUN useradd -m myuser
USER myuser

# Expose the port the app runs on
ARG APP_PORT=5000
EXPOSE $APP_PORT

# Run the Gunicorn server, using the environment variable for the port
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:${APP_PORT}", "run:app"]
