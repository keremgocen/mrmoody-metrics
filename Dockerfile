FROM scratch
CMD ["apt-get install -y ca-certificates"]
ADD mrmoody /app
ADD config.json /
CMD ["./app"]