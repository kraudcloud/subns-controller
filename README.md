a k8s controller that lets an owner of a namespace create more namespaces.

this is particular useful if you have several teams using k8s,
and you want to allow a team to create and delete their own namespaces,
without giving them permissions to accidently break someone elses stuff.

this is for prevening _accidental_ namespace escape.
i have not actually checked if this allows _intentional_ escape,
i.e. don't use this when you dont trust your users.


example:

```yaml
apiVersion: subns.subns.kraud.cloud/v1alpha1
kind: SubNamespaceClaim
metadata:
  name: sub1
  namespace: parent
spec:
  name: sub1
  roleBindings:
    - name: developers
      subjects:
      - kind: ServiceAccount
        name: bobi
        namespace: klum
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: admin

```

this will create an ns called parent-sub1


# usage

Install the CRD:

    make install


deploy the image:

    kubectl apply -f prod.yaml



# dev

build img:

    make docker-build docker-push IMG=ctr.0x.pt/subns-controller/subns-controller:latest
