#  Ephemeral stack administration

*ephstack* toolset provisions & configures application ephemeral stacks on public & private clouds.  

An ephstack are:
- not meant to live long
- meant for launching *non production* workloads like internal demos/POC of a workflow/burst loads for Big data jobs/Software Testing 
- provisioned in a single virtual network (like VPC/Azure VNet etc) i.e 1 tier only. multi tier subnets are not supported.
