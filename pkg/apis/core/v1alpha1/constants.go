package v1alpha1

// SecretReplicationNamespacesAnnotation is the name of the annoation that defines the namespaces where the annotated resource should be replicated to.
const SecretReplicationNamespacesAnnotation = "replication.schrodit.tech/namespaces"

// SecretReplicationAllNamespacesAnnotation is the name of the annoation that defines that the annotated resource should be replicated to all namespaces.
const SecretReplicationAllNamespacesAnnotation = "replication.schrodit.tech/all"

// SecretReplicationReplicaOfAnnotation is the name of the annotation that defines the source resource of the current resource.
const SecretReplicationReplicaOfAnnotation = "replication.schrodit.tech/replicaOf"

// SecretReplicationLasObservedGenerationAnnotation is the name of the annotation that defines the last observed generation of the replicating secret.
const SecretReplicationLasObservedGenerationAnnotation = "replication.schrodit.tech/lastObservedGeneration"
