import { Template } from 'e2b'

export const template = Template()
  .fromImage('e2bdev/base')

  // Switch to root to install packages
  .runCmd('sudo apt-get update')
  .runCmd('sudo apt-get install -y openjdk-17-jdk golang')
  .runCmd('sudo apt-get clean')

  // Verify installations
  .runCmd('java -version')
  .runCmd('go version')
