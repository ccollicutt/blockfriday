# blockfriday

This is a insanely simple admission controller that will block NEW deployments from being created on a Friday. Yes, Friday is hard coded! The point of this isn't to be useful, but to be a fun experiment in admission controllers.

Is it Friday? Let's find out!

```
func isFriday() bool {
    return time.Now().Weekday() == time.Friday
}
```

## Certificates, CA Bundles, Oh My!

The point of all this is to have the Kubernetes API be able to validate the webhook certificate.

To be honest the certificates might be the hardest part of this. I'm using cert-manager to manage the certificates, so if you want to follow this exact process, you'll need to deploy that first.

This example assumes that Kubernetes was created with kubeadm (big assumption) and thus kubeadm created the CA for the control plane which lives in `/etc/kubernetes/pki/`. There's a job that will create a secret from the cert and key files in that directory which will then be used as part of the cert-manager cluster issuer. (Would one do this in the real world, probably not.)

Further, cert-manager has a handy-dandy injector which the validating webhook configuration can use to inject the CA bundle into the webhook configuration. This is done by adding the following annotation to the validating webhook configuration:

```
  annotations:
	cert-manager.io/inject-ca-from: blockfriday/admission-controller-certificate
```

Basic steps:

* Deploy cert-manager
* Create a secret from the kubeadm created certificates

>NOTE: This is only going to work without modification in a kubeadm cluster.

>NOTE: I alias kubectl to k.

```
k create -f cert-manager-setup/create-secret.yaml
```

* Create a cluster issuer

```
k create -f cert-manager-setup/cluster-issuer.yaml
```

## Installation of the Validating Webhook

### Build the Image

Review the makefile and perhaps Dockerfile and run:

```
make image
```

### Deployment

Now that the certificate infrastructure is sorted, we can deploy the webhook.

* Create a namespace for the webhook

```
k create -f manifests/namespace.yaml
```

* Create a certificate for the webhook

```
k create -f manifests/certs.yaml
```

* Create the validating webhook configuration

```
k create -f manifests/validating-webhook-configuration.yaml
```

* Deploy the webhook

```
k create -f manifests/deployment.yaml
k create -f manifests/service.yaml
```

* Try to create a NEW deployment (if it's Friday it will fail)
* Rejoice!

## Yes, Friday is Hardcoded

I thought briefly about making the "blockday" a variable, but then it was going to get too complex because I should probably allow multiple days, then should allow stretches of time, and multiple stretches of time, etc. Then I’d have to error check that config. It was going to be too complicated to do for a first round, so I left it just being Friday.
