#  Ephemeral stack administration

Ephemeral stacks in this project makes the following assumptions:
- Each stack is a combination of infrastructure & apps that are not meant to live long. 
- Each stack is being deployed to run *non production* workloads like internal demos/POC of a workflow/burst loads for Big data jobs/Software Testing 
- Each stack is provisioned in a single virtual network (like VPC/Azure VNet etc) i.e 1 tier only. No multi tier subnets are not supported.
