# pulumi-docker-vps

This repo is a simple example of how to deploy a VPS using Pulumi that has Docker installed and running.

# Usage

1. Create a Pulumi stack
```
pulumi stack init <stack-name>
```

2. Set your DigitalOcean token as a secret:

```
pulumi config set digitalocean:token --secret
```

3. Deploy the stack:

```
pulumi up
```

4. View the stack status for the ip address:

```
pulumi stack output
```

5. SSH into your VPS:

```
ssh root@<ip-address> ./<stack-name>_id_rsa
```

5. Observe the status of the cloud-init script

```
cloud-init status
```

6. Look at logs for issues

```
cat /var/log/cloud-init-output.log
```
