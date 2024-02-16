# Go Kubeseal GUI

A personal project to easily navigate a kubeseal sealed secret

## Features
1. Read existing sealed secret
2. Seal a new/modified secret

## Pre requisites
1. kubectl

2. kubeseal
  https://github.com/bitnami-labs/sealed-secrets

## Installation

```
```

## Configuration

```
```

## Usage

```
```

## Approach
1. This app will interact using your pc command line `os.Run()`
2. It will change the k8s context to your desired context, eg: `kubectl config use-context xxx`

### Read
1. It will execute `kubectl get secret xxx -o yaml`
2. Then it will pipe the result to the golang app then forwarded to the GUI

### Seal
1. It will execute `kubeseal --controller-name=yyy --controller-namespace=zzz --format=yaml -f aaa`
2. Then it will pipe the result to the golang app then forwarded to the GUI
