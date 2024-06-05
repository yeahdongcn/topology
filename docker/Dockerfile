FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt update && apt install -y wget infiniband-diags slurmctld gawk
COPY slurm.conf /etc/slurm-llnl/slurm.conf
COPY slurmibtopology.sh /usr/local/bin/slurmibtopology.sh
RUN chmod +x /usr/local/bin/slurmibtopology.sh

ENTRYPOINT [ "slurmibtopology.sh", "-c" ]