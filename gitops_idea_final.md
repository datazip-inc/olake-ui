# Idea for GitOps support in OLake

## Overview
We are gonna use GitOps as a way in OLake to manage the configuration of the platform (Configuration as Code).

### How it works:
1. User can manage `source`, `destination`, `job`, and `streams` in GitHub as a K8s resource (CRDs).
2. On push, we make use of ArgoCD and build an Operator/Controller whatever is necessary for our use case. 
3. We receive the changes and create/update/delete the olake-ui resources and databases automatically in the Operator. User need not touch UI anymore, all managed via GitHub.
4. The order Source -> Destination -> Job -> Streams added to the Jobs table. 

## Resources:
We are planning to use the same format as we have in the API Contract for Source, Destination and Job CRUD operations. That's how the user will store the details in the GitHub. 

Here's an example that I have:

```yml
apiVersion: v1
kind: ConfigMap # We will create a CRD instead
metadata:
  name: olake-source-1
  labels:
    olake.io/managed-by: gitops # these are exaemples - not sure if we need these labels. 
    olake.io/resource-type: source
    olake.io/project-id: "123"
  annotations:
    argocd.argoproj.io/sync-wave: "0" # Not sure if we need this (just an example - let's follow the most optimal/standard way)
data: # here I have used source config directly - but we can use the same format as in the API contract.
  config.json: |
    {
      "name": "smoke-source",
      "type": "POSTGRES",
      "version": "latest",
      "config": {
        "host": "example.invalid",
        "port": 5432,
        "database": "smoke",
        "username": "smoke",
        "password": "smoke"
      }
    }
```

Note: the order of CRUD matters, I think it is already handled in the endpoints that we have as of today (just make sure)
- Source/destination should be created before Job and Job before streams. 

We will be considering the name of the file or an annotation as the key of each entity. 
Let's say the file name is olake_source___<source_name>.json, olake_destination___<dest_name>.json, olake_job___<job_name>.json or we can use `olake.io/resource-type: source` metadata for it. We can discuss which is the best practice and standard way of doing it. 

Now in the Jobs API Contract:
```json
type CreateJobRequest struct {
	Name             string            `json:"name" binding:"required" example:"my-sync-job"`
	Source           *DriverConfig     `json:"source" binding:"required"`
	Destination      *DriverConfig     `json:"destination" binding:"required"`
	Frequency        string            `json:"frequency" binding:"required" example:"0 */6 * * *"`
	StreamsConfig    string            `json:"streams_config" orm:"type(jsonb)" binding:"required"`
	Activate         bool              `json:"activate,omitempty" example:"true"`
	AdvancedSettings *AdvancedSettings `json:"advanced_settings,omitempty"`
}

type UpdateJobRequest struct {
	Name              string            `json:"name" binding:"required" example:"my-sync-job-updated"`
	Source            *DriverConfig     `json:"source" binding:"required"`
	Destination       *DriverConfig     `json:"destination" binding:"required"`
	Frequency         string            `json:"frequency" binding:"required" example:"0 */12 * * *"`
	StreamsConfig     string            `json:"streams_config" orm:"type(jsonb)" binding:"required"`
	DifferenceStreams string            `json:"difference_streams,omitempty" example:"[]"`
	Activate          bool              `json:"activate,omitempty" example:"true"`
	AdvancedSettings  *AdvancedSettings `json:"advanced_settings,omitempty"`
}
```

We have SourceConfig, DestinationConfig, StreamsConfig, DifferenceStreams. We already have them mentioned in other separate files. We cannot mention them here again. So we can think of a better approach. Here's my idea - we can set the name of the entity here, so that we can in the watcher/controller we have we can fetch from the database by name and attach it here and call the API we have to handle the request. 


IMPORTANT CASE: 
The Update Job is a little complex and robust. 
- On Job Creation, we create the Job the job with the streams.json which comes from another configMap. 
- On Job Update, we first send the streams.json that we have in the Github - and send it to the streams difference API - we get response and send the new streams and the streamDifference in the Update API call. (We can check and discuss for edge cases here if there's any)
  

--- 


## Things to discuss
- There as some things I have mentioned to be discussed, we need to think about those. 
- Then we need to check how are we gonna do error handling - like if source is created, destination is created and Job creation fails. If destination is failed and in Job creation since there's no Job we need to fail as well right. And how is the user notified something has went wrong. 
- We need to handle idempotency as well. 



## Things decided and needs confirmation if it's feasible and if it's a standard way to do it:
- In ArgoCD we can fail the sync status as not sycned properly if any of the operation in the Opereator fails - for example Database save fails, source/dest test connection fails or any issue in the Operator. 
- We need to show in the ArgoCD dashboard that something wrong has happened. 
  

## Things we need to decide:
- Is one CRD enough for (source, destination, job and streams)
- how are we gonna handle failure. 
- we will be getting the body as it is in the API contract of source, dest, and job. They need to enter the ID, and all the details as it is. for streams alone we can handle a separate file as it's very big and attach it with the Job in the operator. We need to decide how we are gonna map multiple sources, dests and jobs. 

Basically we need a complete end-to-end plan and discuss things were we aren't clear about. 
