# Troubleshooter-operator

## Motivation
We'd like to expose some kind of POD that has all the ingredients to troubleshoot cluster wide resources, like networking connections, validate certificates and so on. 
We could utilize kubectl run to create a temp pod, the downside of this way, that it is not a unified way of starting with troubleshooting. 
Some issues regarding this way of troubleshooting arise: 
- it needs the so called God-mode / RunAsRoot configurations. 
- the run command can be adjusted with run commands (adding mem/cpu and so on) (not unified)
- the pod needs to be removed after the troubleshooting session is over
- only the cluster-admins can create this pod in the namespaces that are excluded from Gatekeepers constraints. 
- troubleshooting session is not logged (commands, can be copy/pasted to local and shared with team)

These questions can be answered by introducing an operator which lets us create an unified way of troubleshooting. The troubleshooter-operator can be configured in a way that it will listen to a CRD with common sense defaults. As well as prep a POD with all the configurations it needs, like: 
- runAsNonRoot, everyone can deploy the troubleshooter pod.
- optional: Get client-certificates to communicate to a remote-vm (ssh)
- remove itself when inactive or scheduled for a period of time
- log history / output of commands to disk > a PersistenVolume, Creates a PVC, reclaimPolicy retain. 
  - utilize a StorageClass and writes output of troubleshooting session to disk using a PVC. 

## Goals
The goal is ot get an unified way of starting a troubleshooter session, with the option to log the output of the session to disk. 

## Non-goals 
The way how to troubleshoot is out of scope.

## Proposal
Utilize the KUBEBUILDER project for scaffolding the operator / controller-manager way of working. 
In short, we'll create CRDs, like `troubleshooter.sh` which contain `common sense` defaults to stop/start/save troubleshooting sessions, which will be reconciled. 
The CRDs contain at the minimum: 
- name > name of the pod
- image (like ubuntu... ) > this lets us prepare a container with the proper rights/all packages already included. 
- timeToLive > the amount of time that the pod should be UP and running. With a maximum of 2hr? (UTC)
- saveToDisk > setup PVC's to save the troubleshooter session to disk; this needs a PV with ReclaimSetting Retain, to get the logs. 
  - name of PVC 
    - Perhaps log to STDOUT > let something like filebeat pick the logs up. 
- [optional] runAsNonRoot > default is non-root (if this can be managed by setting up container)
- [optional] runAsUser > the link the container with a user (set sudo privileges etc.)
- METRICS!

- Optional > create a temporary namespace (incident-response/uid)

## Design


## Implementation

- record all activities during troubleshooting
  - Simple way is to run the BASH commanbd `script -a my_terminal_activities` during the beginning of the container, and when the container stops, `exit` the script, and make sure the file is saved (`mounted to a volume`)

- container pckg:
  - apt install iproute2 (`ip`)
  - apt-get install net-tools (`ifconfig`)
  - apt-get install dnsutils -y (`nslookup`)
  - apt-get install traceroute -y (`traceroute`) 
  - apt-get install net-tools (`netstat`)
  - apt-get install openssl (`openssl`)
  - apt-get install curl -y (`curl`)

unknowns: 
how to's? 
- write stdout/stderr to disk. 