# RNNR: distributed task execution system for scaling reproducible workflows

*RNNR* is a bioinformatics task processing system for distributed computing environments.
It was designed to work with a workflow management system, such as [Cromwell](https://github.com/broadinstitute/cromwell) and [Nextflow](https://www.nextflow.io/), through the [Task Execution Service (TES) API](https://github.com/ga4gh/task-execution-schemas).
The workflow manager submits tasks to RNNR that distributes the processing load among computational nodes connected in a local network.
The system is composed of the *master* instance and one or more *worker* instances.
A distributed file system (NFS for example) is required.

Full documentation at <https://bcblab.org/rnnr>.
