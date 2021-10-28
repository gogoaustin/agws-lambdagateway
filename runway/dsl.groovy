import utilities.GogoUtilities

SHELL_STEPS='''
mkdir -p `pwd`/.go
export GOPATH=`pwd`/.go
mkdir -p $GOPATH/bin
export PATH=$PATH:$GOPATH/bin

echo "Downloading dep"
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

mkdir -p runway/FS_ROOT/opt
cp -R repo/runway/. runway/
cp repo/build.gradle .
mkdir -p $GOPATH/src/git.gogoair.com/bagws/lambdagateway
cp -r repo/* $GOPATH/src/git.gogoair.com/bagws/lambdagateway/
cp repo/.env $GOPATH/src/git.gogoair.com/bagws/lambdagateway/
cd $GOPATH/src/git.gogoair.com/bagws/lambdagateway
dep ensure

go build -o ./bin/lambdagateway main.go

rm -rf /tmp/workspace/a_bagws_lambdagateway/runway/FS_ROOT/opt/lambdagateway
mkdir -p /tmp/workspace/a_bagws_lambdagateway/runway/FS_ROOT/opt/lambdagateway/
cp .env /tmp/workspace/a_bagws_lambdagateway/runway/FS_ROOT/opt/lambdagateway/
cp bin/lambdagateway /tmp/workspace/a_bagws_lambdagateway/runway/FS_ROOT/opt/lambdagateway/
echo "🎉  Build Process Complete!  🎉"
'''

def myJob = job("$SRC_JOB") {
  parameters {
    stringParam('GIT_BUILD_BRANCH', 'master', 'Git branch used to build.')
  }
  logRotator {
    numToKeep(-1)
    artifactNumToKeep(-1)
  }
  wrappers {
    preBuildCleanup()
  }
  scm {
    git {
      remote {
        url("git@git.gogoair.com:bagws/lambdagateway.git")
      }
      extensions {
        wipeOutWorkspace()
        relativeTargetDirectory("repo")
      }
      branch('$GIT_BUILD_BRANCH')
    }
  }
  steps {
    shell(SHELL_STEPS)
  }
}

g = new GogoUtilities(job: myJob).addBaseOptions()
g.addGradleSteps('rpm')
g.addSlack(slack_room='#gws-dev')