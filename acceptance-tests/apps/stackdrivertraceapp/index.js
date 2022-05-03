
const port = process.env.PORT || 8080

async function main () {
  const vcapServices = require('vcap_services')
  const credentials = vcapServices.findCredentials({ instance: { tags: 'tracing' } })
  if (typeof credentials !== 'object' || Object.entries(credentials).length === 0) {
    throw new Error('could not find credentials in VCAP_SERVICES')
  }

  require('@google-cloud/trace-agent').start({
    logLevel: 4,
    enabled: true,
    projectId: credentials.ProjectId,
    bufferSize: 1,
    credentials: JSON.parse(Buffer.from(credentials.PrivateKeyData, 'base64').toString('ascii'))
  })
  const tracer = require('@google-cloud/trace-agent').get()

  const express = require('express')
  const helmet = require('helmet')
  const app = express()
  app.use(helmet())
  app.use(express.text({ limit: '1kb', type: '*/*' }))
  app.get('/', handleRequest(tracer))
  app.get('/:spanName', handleCustomSpanRequest(tracer))

  app.listen(port, () => console.log(`listening on port ${port}`))
}

const handleCustomSpanRequest = (tracer) => async (req, res) => {
  try {
    console.log('Handling request')
    const spanName = req.params.spanName

    console.log('Creating custom profiling')
    const customSpan = tracer.createChildSpan({ name: spanName })

    res.send({ ProjectId: tracer.getWriterProjectId(), TraceId: tracer.getCurrentContextId() })

    customSpan.endSpan()

    console.log('Finished custom profiling')
  } catch (e) {
    res.status(500).send(e)
  }
}

const handleRequest = (tracer) => async (req, res) => {
  try {
    console.log('Handling request')
    res.send({ projectID: tracer.getWriterProjectId(), tracerConfig: tracer.getConfig() })
  } catch (e) {
    res.status(500).send(e)
  }
}

(async () => {
  try {
    await main()
  } catch (e) {
    console.error(`failed: ${e}`)
  }
})()
