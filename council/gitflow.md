# Git flow

## The main branches

At the core, the development model is greatly inspired by existing models out there. 
The central repo holds two main branches with an infinite lifetime:

* master (which will be renamed as `main`)
* develop

We consider origin/develop to be the main branch where the source code of HEAD always 
reflects a state with the latest delivered development changes for the next release. 
Some would call this the “integration branch”. This is where any automatic nightly 
builds are built from.

When the source code in the develop branch reaches a stable point and is ready to be 
released, all of the changes should be merged back into master somehow and then tagged 
with a release number. How this is done in detail will be discussed further on.

Therefore, each time when changes are merged back into master, this is a new production 
release by definition. 

![Main branches](img/main-branches.png "Main branches")

## Supporting branches

Next to the main branches master and develop, our development model uses a variety of 
supporting branches to aid parallel development between team members, ease tracking of 
features, prepare for production releases and to assist in quickly fixing live production 
problems. Unlike the main branches, these branches always have a limited life time, 
since they will be removed eventually.

The different types of branches we may use are:

* Feature branches
* Release branches
* Hotfix branches

Each of these branches have a specific purpose and are bound to strict rules as to which 
branches may be their originating branch and which branches must be their merge targets. We 
will walk through them in a minute.

By no means are these branches “special” from a technical perspective. The branch types are 
categorized by how we use them. They are of course plain old Git branches.

## Feature branches

Feature branches (or sometimes called topic branches) are used to develop new features for 
the upcoming or a distant future release. When starting development of a feature, the target 
release in which this feature will be incorporated may well be unknown at that point. The 
essence of a feature branch is that it exists as long as the feature is in development, 
but will eventually be merged back into develop (to definitely add the new feature to 
the upcoming release) or discarded (in case of a disappointing experiment).

![Feature branches](img/feature-branch.png "Feature branch")

Feature branches should always branch off from `develop` branch and must merge back into
`develop` too. The branch naming convention for feature branches should be anything like
`feature/something`, where something can be anything except master, develop, release-*, 
or hotfix-*.

To create a feature branch:

```bash
# certify you are at the develop branch
$ git status                       
On branch develop

# update branch
$ git fetch --all --prune
$ git pull origin develop

# create new feature/branch
$ git checkout -b feature/awesome-feature

# edit new code, add and commit (several time, short commits)
$ git commit -m "Awesome commit message"

# push new branch to github
$ git push origin feature/awesome-feature
```

When new feature is done you should open a pull request pointing to `develop` to 
merge changes at the *central repository* (aka origin). After merge delete the branch
at origin using Github interface. 

### When is my branch ready to be merged?

After sending the new code written in the form of a pull request, it must be reviewed 
by the team, who will make comments that they consider relevant as:

* doubts about syntax
* potential problems not previously identified
* typos and other problems

If the reviewer considers that the code sent is ok he will approve the pull request, 
indicating that in his evaluation the new code is ready to join the `develop` branch.

The code reviewer should keep in mind that he becomes co-responsible for the written 
code at the time of its approval. Therefore, the review must be done with great care 
and attention. Also, the person who created the PR is responsible for doing the merge 
after having the approval of at least one reviewer.

## Hotfix branches

Hotfix branches are very much like release branches in that they are also meant to 
prepare for a new production release, albeit unplanned. They arise from the necessity 
to act immediately upon an undesired state of a live production version. When a 
critical bug in a production version must be resolved immediately, a hotfix branch may 
be branched off from the corresponding tag on the master branch that marks the 
production version.

![Hotfix branches](img/hotfix-branches.png "Hotfix branches")

The essence is that work of team members (on the develop branch) can continue, while 
another person is preparing a quick production fix.

Hotfix branches are created from the master branch. For example, say version 1.2 is 
the current production release running live and causing troubles due to a severe bug. 
But changes on develop are yet unstable. We may then branch off a hotfix branch and 
start fixing the problem:

```bash
# certify you are at the master branch
$ git status                       
On branch master

# update branch
$ git fetch --all --prune
$ git pull origin master

# create hotfix branch
$ git checkout -b hotfix/1.2.1

# Update and add files, bump version and commit
git commit -m "Fix something. Bumped version number to 1.2.1"

# push new branch and open a pull request pointing to master
$ git push origin hotfix/1.2.1
```

After merge with *master* you need to update the *develop* branch with the hotfix too
to avoid conflicts and loose code. You should create a new PR that will merge the 
hotfix branch with *develop* branch. This must be approved too. 

Hotfix branches should always branch off from `master` branch and must merge back into
`master` too. Also this should be merged in `develop` too. The branch naming convention 
for hotfix branches should be anything like `hotfix/something`, where something can be 
anything except master, develop, release-*, or hotfix-*.

## Release branches

Release branches support preparation of a new production release. They allow for 
last-minute dotting of i’s and crossing t’s. Furthermore, they allow for minor bug fixes 
and preparing meta-data for a release (version number, build dates, etc.). By doing all 
of this work on a release branch, the develop branch is cleared to receive features for 
the next big release.

The key moment to branch off a new release branch from develop is when develop (almost) 
reflects the desired state of the new release. At least all features that are targeted 
for the release-to-be-built must be merged in to develop at this point in time. All 
features targeted at future releases may not—they must wait until after the release 
branch is branched off.

It is exactly at the start of a release branch that the upcoming release gets assigned 
a version number—not any earlier. Up until that moment, the develop branch reflected 
changes for the “next release”. Each new commit in *develop* must be named xx.xx.xx-rcx,
example: v1.2.0-rc5. Subsequent commits will increase the X in -rcX, although maybe not 
all commits must include the tag. At the end of the day those tags are used to indicate 
the QA team that "this code should be tested to check if it's production ready".

![Release model](img/release-branch.png "Release model")

Release branches are created from the develop branch. For example, say version 
1.1.5 is the current production release and we have a big release coming up. The 
state of develop is ready for the “next release” and we have decided that this will 
become version 1.2 (rather than 1.1.6 or 2.0). So we branch off and give the release 
branch a name reflecting the new version number:

```bash
# certify you are at the develop branch
$ git status                       
On branch develop

# update branch
$ git fetch --all --prune
$ git pull origin develop

# create release branch
$ git checkout -b release/name-of-release
Switched to a new branch "release/name-of-release"
```

After that open a PR pointing to `master` to put the release branch inside the 
approval flow. After approved the responsible for open the PR must merge into `master`,
delete the release branch and **tag new version**. 

Since *master* is a protected branch, not all users are enable to merge into this. So
if you don't have access to do this please ask somebody able to do after approvals and
all flow be satisfied. 

## References

This document was written using this [link](https://nvie.com/posts/a-successful-git-branching-model/#supporting-branches)
as reference. 


