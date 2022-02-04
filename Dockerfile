##################################
# STEP 1 build executable binary #
##################################
FROM golang:1.16-alpine AS builder

# Install git. Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

ENV APP_HOME /go/src/github.com/nobled5/modbus_exporter
WORKDIR $APP_HOME

# Copy src.
COPY  --chown=nobody:nogroup   /cmd/modbus_exporter/  $APP_HOME/cmd/modbus_exporter/
COPY  --chown=nobody:nogroup   /pkg/collector/        $APP_HOME/pkg/collector/
COPY  --chown=nobody:nogroup   /pkg/config/           $APP_HOME/pkg/config/
COPY  --chown=nobody:nogroup   /pkg/handler/          $APP_HOME/pkg/handler/
COPY  --chown=nobody:nogroup   /pkg/logger/           $APP_HOME/pkg/logger/
COPY  --chown=nobody:nogroup   /pkg/master/           $APP_HOME/pkg/master/
COPY  --chown=nobody:nogroup   /pkg/resources/        $APP_HOME/pkg/resources/
COPY  --chown=nobody:nogroup   /pkg/structures/       $APP_HOME/pkg/structures/
COPY  --chown=nobody:nogroup   /pkg/workload/         $APP_HOME/pkg/workload/

COPY  /go.mod                  $APP_HOME/

COPY  --chown=nobody:nogroup   /pkg/resources/        /etc/modbus_exporter/resources/
COPY  --chown=nobody:nogroup   modbus.yaml            /etc/modbus_exporter/

# Fetch dependencies using go get.
RUN go get -d -v $APP_HOME/cmd/modbus_exporter/

ARG ver="0.0.1"
ARG branch="HEAD"
ARG hash="hash"
ARG user="nobled5"
ARG date="20060102-15:04:05"

ENV VERSION   "-X github.com/prometheus/common/version.Version=$ver"
ENV BRANCH    "-X github.com/prometheus/common/version.Branch=$branch"
ENV REVISION  "-X github.com/prometheus/common/version.Revision=$hash"
ENV USER      "-X github.com/prometheus/common/version.BuildUser=$user"
ENV DATE      "-X github.com/prometheus/common/version.BuildDate=$date"

# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build \
  -ldflags="-w -s $VERSION $BRANCH $REVISION $USER $DATE" \
  -o /go/bin/main /go/src/github.com/nobled5/modbus_exporter/cmd/modbus_exporter/

##############################
# STEP 2 build a small image #
##############################
FROM scratch

LABEL description="Scratch-based Docker image for modbus_exporter"

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd    /etc/passwd

# Copy static executable.
COPY --from=builder /go/bin/main   /bin/modbus_exporter

# Copy yaml conf.
COPY --from=builder /etc/modbus_exporter/resources/  /resources

# Copy yaml conf.
COPY --from=builder /etc/modbus_exporter/modbus.yaml /etc/modbus_exporter/modbus.yaml

# Use an nobody user.
USER nobody

VOLUME     [ "/modbus_exporter" ]
WORKDIR     /modbus_exporter

# Run binary.
ENTRYPOINT [ "/bin/modbus_exporter" ]
CMD        [ "--config.file=/etc/modbus_exporter/modbus.yaml" ]
