# EC2 Deployment Guide with Terraform

## Overview
This Terraform configuration will:
1. Create an EC2 instance (t2.micro) on Amazon Linux 2023
2. Install Git and Docker automatically
3. Clone your GitHub repository
4. Build and run your Docker container
5. Expose port 8080 for web access

## Prerequisites
- AWS credentials configured (AWS Academy/AWS CLI)
- An existing AWS key pair
- Your local IP address for SSH access

## Setup Steps

### 1. Configure Variables
Edit `terraform.tfvars` with your values:
```hcl
ssh_cidr     = "YOUR_IP/32"    # Get your IP from https://whatismyip.com
ssh_key_name = "YOUR_KEY_NAME"  # Your AWS key pair name (e.g., "vockey")
```

### 2. Initialize Terraform
```bash
cd HW-1/terraform
terraform init
```

### 3. Review the Plan
```bash
terraform plan
```

### 4. Deploy
```bash
terraform apply
```
Type `yes` when prompted.

### 5. Get Outputs
After deployment, Terraform will output:
- `ec2_public_ip` - Use this to access your service
- `ec2_public_dns` - DNS name of your instance

### 6. Test Your Service
Wait 2-3 minutes for the EC2 instance to complete initialization, then:

```bash
# Test with curl
curl http://YOUR_EC2_PUBLIC_IP:8080/albums

# Or open in browser
http://YOUR_EC2_PUBLIC_IP:8080/albums
```

You should see the albums JSON response!

## Architecture Details

### Security Group Rules
- **Port 22 (SSH)**: Restricted to your IP address
- **Port 8080 (Web App)**: Open to the internet (0.0.0.0/0)
- **Egress**: All outbound traffic allowed

### User Data Script
The EC2 instance runs this script on first boot:
```bash
#!/bin/bash
yum update -y
yum install -y git docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Clone and Run App
git clone https://github.com/justin-aj/go-hw1.git /home/ec2-user/app
cd /home/ec2-user/app/web-service-gin
docker build -t web-service-gin .
docker run -d -p 8080:8080 web-service-gin
```

## Troubleshooting

### If the service isn't responding:
1. SSH into the instance:
   ```bash
   ssh -i your-key.pem ec2-user@YOUR_EC2_PUBLIC_IP
   ```

2. Check Docker container status:
   ```bash
   sudo docker ps
   sudo docker logs <container_id>
   ```

3. Check user data script execution:
   ```bash
   sudo cat /var/log/cloud-init-output.log
   ```

### Update the Git Repository URL
If your GitHub repo is different, update line 31 in `main.tf`:
```hcl
git clone https://github.com/YOUR_USERNAME/YOUR_REPO.git /home/ec2-user/app
```

## Cleanup
When done, destroy resources to avoid charges:
```bash
terraform destroy
```
Type `yes` when prompted.

## Cost Note
t2.micro instances are eligible for AWS Free Tier (750 hours/month).
