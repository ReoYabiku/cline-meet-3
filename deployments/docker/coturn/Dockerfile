FROM coturn/coturn:latest

# Copy custom configuration
COPY deployments/docker/coturn/turnserver.conf /etc/coturn/turnserver.conf

# Create necessary directories and set permissions
USER root
RUN mkdir -p /var/log/coturn /var/run && \
    chmod 755 /var/log/coturn /var/run && \
    chmod 644 /etc/coturn/turnserver.conf

# Expose STUN/TURN ports
EXPOSE 3478/udp 3478/tcp
EXPOSE 5349/udp 5349/tcp
EXPOSE 49152-65535/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD turnutils_stunclient -p 3478 127.0.0.1 || exit 1

# Start coturn as root (simpler for development)
CMD ["turnserver", "-c", "/etc/coturn/turnserver.conf"]
