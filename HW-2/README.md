# HW-2: AWS EC2 Infrastructure Setup

Setting up and managing cloud infrastructure on AWS using Terraform. This assignment demonstrates Infrastructure as Code (IaC) principles and cloud resource management.

## Overview

This project automates the provisioning of AWS EC2 instances and related infrastructure using Terraform. It covers infrastructure definition, state management, and cloud operations.

## Project Structure

```
HW-2/
├── main.tf                    # Terraform configuration for EC2, AMI, security groups
├── terraform.tfvars          # Variable values (secrets, configuration)
├── terraform.tfstate         # Current infrastructure state
├── terraform.tfstate.backup  # Backup of previous state
├── hw2.pem                   # SSH private key for EC2 access
├── .terraform/               # Terraform internal directory
├── .terraform.lock.hcl       # Dependency lock file

```

## Prerequisites

- Terraform 1.0 or higher
- AWS CLI configured with credentials
- AWS Account with appropriate permissions
- SSH client for accessing EC2 instances

## Configuration

### Variables (terraform.tfvars)

Key variables to configure:
- `ssh_cidr`: Your home IP in CIDR notation (e.g., `203.0.113.45/32`)
- `ssh_key_name`: Name of your existing AWS key pair

### AWS Setup Requirements

1. Create an AWS key pair for SSH access:
   ```bash
   aws ec2 create-key-pair --key-name <key-name> --region us-east-1
   ```

2. Create `terraform.tfvars` with your configuration:
   ```hcl
   ssh_cidr      = "YOUR_IP/32"
   ssh_key_name  = "YOUR_KEY_PAIR_NAME"
   ```

## Terraform Commands

### Initialize Terraform
```bash
terraform init
```

### Plan Infrastructure Changes
```bash
terraform plan
```

### Apply Configuration
```bash
terraform apply
```

### View Current State
```bash
terraform state list
terraform state show aws_instance.demo-instance
```

### Destroy Infrastructure
```bash
terraform destroy
```

## Infrastructure Components

### EC2 Instance
- **AMI**: Amazon Linux 2023 (AL2023)
- **Instance Type**: Configurable (typically t2.micro for free tier)
- **Region**: us-east-1
- **Security**: SSH access restricted to your IP only

### Security Groups
- Inbound: SSH (port 22) from your IP only
- Outbound: All traffic allowed
- Protects against unauthorized access

### Network Configuration
- VPC: Default VPC
- Subnet: Default subnet
- Public IP: Assigned for internet access

## Connecting to Your Instance

Once infrastructure is deployed:

```bash
ssh -i hw2.pem ec2-user@<INSTANCE_PUBLIC_IP>
```

Replace `<INSTANCE_PUBLIC_IP>` with the actual IP from AWS console or Terraform output.

## Key Concepts

### Infrastructure as Code (IaC)
- Version control for infrastructure
- Reproducible deployments
- Disaster recovery
- Team collaboration

### Terraform State Management
- **terraform.tfstate**: Current infrastructure state
- **terraform.tfstate.backup**: Previous state backup
- Keep state files secure and backed up
- Consider remote state for team environments

### Security Best Practices
- Restrict SSH access to your IP only
- Use key pairs instead of passwords
- Keep `terraform.tfvars` with secrets out of version control
- Regularly rotate AWS credentials
- Use AWS IAM for fine-grained permissions

## Common Tasks

### Modify Instance Type
Edit `main.tf` and change `instance_type`:
```hcl
instance_type = "t2.small"
```
Then run:
```bash
terraform plan
terraform apply
```

### Add Elastic IP
```hcl
resource "aws_eip" "demo_eip" {
  instance = aws_instance.demo-instance.id
  domain   = "vpc"
}
```

### Create Additional Instances
Duplicate and modify the EC2 resource or use Terraform modules.

## Troubleshooting

### Permission Denied Connecting via SSH
- Verify SSH key permissions: `chmod 400 hw2.pem`
- Ensure security group allows port 22 from your IP
- Check instance is in "running" state

### Terraform Apply Fails
- Verify AWS credentials are configured
- Check IAM permissions
- Review error messages for resource conflicts

### State Lock Issues
```bash
terraform force-unlock <LOCK_ID>
```

## Best Practices

1. **Always run `terraform plan` before `apply`**
   - Review changes before applying
   - Prevents accidental deletion

2. **Keep State Files Secure**
   - Use S3 remote backend for team environments
   - Enable versioning and encryption

3. **Use Variables for Flexibility**
   - Separate configuration from code
   - Different values per environment

4. **Tag Resources**
   - Add tags for cost tracking
   - Organize and identify resources

5. **Monitor Costs**
   - Use AWS Cost Explorer
   - Set up billing alerts
   - Clean up unused resources

## References

- [Terraform Documentation](https://www.terraform.io/docs)
- [AWS Provider Reference](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [EC2 User Guide](https://docs.aws.amazon.com/ec2/)
- [AWS Security Best Practices](https://docs.aws.amazon.com/security/)

## Notes

- Free tier eligible: t2.micro instance with minimal usage
- Always destroy resources when not in use to avoid charges
- Monitor AWS billing dashboard
- Terraform state files contain sensitive information - protect them
