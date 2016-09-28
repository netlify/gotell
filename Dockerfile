FROM calavera/go-glide:v0.12.2

ADD . /go/src/github.com/netlify/netlify-comments

RUN useradd -m netlify && cd /go/src/github.com/netlify/netlify-comments && make deps build && mv netlify-comments /usr/local/bin/

USER netlify
CMD ["netlify-comments"]
