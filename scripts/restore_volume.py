
import argparse
from kubernetes import client, config

config.load_kube_config()
core_api = client.CoreV1Api()
apps_api = client.AppsV1Api()
cobj_api = client.CustomObjectsApi()

def set_replica_count(deployment, replica_count):
    pass

def restore_volume(pvc):
    pass

def get_pvcs(namespace):
    pass

def get_volume_snapshots(namespace):
    snapshots = []

    for snap in cobj_api.list_namespaced_custom_object("snapshot.storage.k8s.io", "v1", namespace, "volumesnapshots")["items"]:
        snapshots.append({
            "name": snap["metadata"]["name"],
            "pvc": snap["spec"]["source"]["persistentVolumeClaimName"],
            "creationTime": snap["status"]["creationTime"]
        })

    return snapshots

if __name__ == "__main__":
    # 1. namespace is cli argument
    namespace = "prowlarr"

    # 2. list all pvcs and let user select which ones to restore

    # 3. scale down all deployments using the pvcs

    # 4. delete pvcs

    # 5. restore pvcs from snapshots

    # 6. scale deployments back up

    # find deployment(s)
    deployments = apps_api.list_namespaced_deployment(namespace)
    # print(deployments)
    deployment = deployments.items[0]
    volumes = deployment.spec.template.spec.volumes
    pvcs = [v.persistent_volume_claim.claim_name for v in volumes if v.persistent_volume_claim is not None]
    # print(pvcs)

    snaps = get_volume_snapshots(namespace)
    print(snaps)




    # print(deployment)
    deployment = apps_api.read_namespaced_deployment("prowlarr", namespace)

    exit(0)

    # find pvc(s) with VolumeSnapshot
    pvc = "prowlarr-config"

    # confirm actions
    # TODO

    # 1. scale down deployment (set replicas to 0)
    set_replica_count(deployment, 0)

    # 2. restore volume from snapshot
    restore_volume(pvc)

    # 3. scale up deployment (reset replicaCount)
    set_replica_count(deployment, 1) # TODO: reset to memorized count

    # print("Listing pods with their IPs:")
    # ret = v1.list_pod_for_all_namespaces(watch=False)
    # for i in ret.items:
    #     print("%s\t%s\t%s" % (i.status.pod_ip, i.metadata.namespace, i.metadata.name))