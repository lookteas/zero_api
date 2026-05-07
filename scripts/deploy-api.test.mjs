import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const localDeploySource = readFileSync(new URL('./deploy-api.ps1', import.meta.url), 'utf8')
const serverDeploySource = readFileSync(new URL('./deploy-api.sh', import.meta.url), 'utf8')
const readme = readFileSync(new URL('../README.md', import.meta.url), 'utf8')

test('deploy-api script builds and uploads a linux binary over ssh', () => {
  assert.equal(localDeploySource.includes('HostName'), true)
  assert.equal(localDeploySource.includes('UserName'), true)
  assert.equal(localDeploySource.includes('RemotePath'), true)
  assert.equal(localDeploySource.includes('ServiceName'), true)
  assert.equal(localDeploySource.includes('GOOS=linux'), true)
  assert.equal(localDeploySource.includes('GOARCH'), true)
  assert.equal(localDeploySource.includes('go test ./...'), true)
  assert.equal(localDeploySource.includes('go build'), true)
  assert.equal(localDeploySource.includes('& scp'), true)
  assert.equal(localDeploySource.includes('& ssh'), true)
})

test('deploy-api script backs up, replaces, restarts, and checks the service', () => {
  assert.equal(localDeploySource.includes('.bak.'), true)
  assert.equal(localDeploySource.includes('systemctl restart'), true)
  assert.equal(localDeploySource.includes('systemctl status'), true)
  assert.equal(localDeploySource.includes('HealthCheckUrl'), true)
  assert.equal(localDeploySource.includes('curl -fsS'), true)
})

test('server deploy script pulls source and restarts systemd service', () => {
  assert.equal(serverDeploySource.includes('git pull --ff-only'), true)
  assert.equal(serverDeploySource.includes('go test ./...'), true)
  assert.equal(serverDeploySource.includes('go run zero.go -f'), true)
  assert.equal(serverDeploySource.includes('systemctl restart'), true)
  assert.equal(serverDeploySource.includes('systemctl status'), true)
  assert.equal(serverDeploySource.includes('ZERO_API_SERVICE'), true)
  assert.equal(serverDeploySource.includes('systemctl cat'), true)
  assert.equal(serverDeploySource.includes('nohup'), false)
  assert.equal(serverDeploySource.includes('http://127.0.0.1:8888/api/v1/health'), true)
  assert.equal(serverDeploySource.includes('curl -fsS'), true)
})

test('readme documents deploy-api usage', () => {
  assert.equal(readme.includes('scripts/deploy-api.sh'), true)
  assert.equal(readme.includes('scripts/deploy-api.ps1'), true)
  assert.equal(readme.includes('ServiceName'), true)
  assert.equal(readme.includes('ZERO_API_CONFIG'), true)
  assert.equal(readme.includes('/api/v1/health'), true)
})
