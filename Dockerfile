FROM scratch
ADD mrmoody-metrics /app
ADD config.json /
CMD ["/app"]