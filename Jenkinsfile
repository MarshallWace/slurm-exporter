import com.mwam.jenkins.build.*

def dockerBuild = new DockerBuild()
dockerBuild.ImageName = "prometheus-slurm-exporter"

automatic_release(dockerBuild, ["master"])