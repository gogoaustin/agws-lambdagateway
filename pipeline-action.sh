#!/bin/bash

####
# Assumptions: This file will be placed in the root of the repository. I.e. gogobiz/BAUDP-xdw
# All the related runway files are under the runway directory
####

#Defined directory & files to search for
DIRECT='runway'
DSL='dsl.groovy'
QE='QE'
APPLICATION='pipeline.json'
REPO=$1

if [ -f $DIRECT/$DSL ] && [ -f $DIRECT/$QE ]; then
	echo 'DO NOTHING'
elif [ -f $DIRECT/$DSL ] && [ ! -f $DIRECT/$QE ]; then
	echo 'FOUND DSL, NO QE'
	curl -vs "https://api:11819cbe5fda7c668998c2808df01da8f7@jenkins-ci.gogoair.cloud/job/dsl-runway/buildWithParameters?token=Hq9vSn2hDK1pjbTbuh4fyq&REPO=$REPO"
	if [ -f $DIRECT/$APPLICATION ]; then
	echo 'FOUND APPLICATION.JSON FILE'
	curl -vs "https://api:11819cbe5fda7c668998c2808df01da8f7@jenkins-ci.gogoair.cloud/job/pipes-app/buildWithParameters?token=khj34uidrfg870weuhw3490eijoert0345jrd0925jdfgio&REPO=$REPO"
    fi
fi
if [ -f $DIRECT/$QE ]; then
	curl -vs "https://api:11d5ca8e9502f45870da83482075972601@jenkins-qe.gogoair.cloud/job/dsl-runway/buildWithParameters?token=khj34uidrfg870weuhw3490eijoert0345jrd0925jdfgio&REPO=$REPO"
fi