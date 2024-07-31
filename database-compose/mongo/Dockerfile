FROM mongo
COPY replica.key /data/replica.key
RUN chmod 400 /data/replica.key
RUN chown 999:999 /data/replica.key
RUN echo "$(cat /data/replica.key)"
