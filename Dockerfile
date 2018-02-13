FROM golang

RUN apt-get update && apt-get install --yes --auto-remove \
		apt-utils \
		curl \
		file \
		git \
		jq \
		nano \
		net-tools \
		tree \
		wget \
	&& rm -rf /var/lib/apt/lists/*

RUN curl https://glide.sh/get | sh
RUN go get -v -u golang.org/x/tools/cmd/goimports
RUN go get -v -u github.com/derekparker/delve/cmd/dlv
RUN go get -v -u github.com/cespare/reflex

ENTRYPOINT bash
CMD ["tail", "-f", "/dev/null"]
