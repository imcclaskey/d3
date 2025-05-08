project-results: rethink string 
switch-feature: switch between features, don't just create them
create-and-define: when creating a feature, we should enter the define phase automatically. 
task: full task support with tool calls
 -- do we really need tool calls? The trial of delivery is workin well without them. We may need some sort of "alterations" field which tracks any divergence from the plan. 
 -- Github interactions should check working directory and stage the files for the user, as well as suggest a  commit message. It shouldn't commit itself though.
github: github support (tasks, commits, etc). Also support gitignore files!
smooth-setup: a setup flow that helps configuration (working dir set, filling out setup/workcheck, etc)
enter-exit: Ability to enter and exit d3, which "turns off" the rules. resume? start? stop? hmmmmm
stateful-features: save your feature state when entering and exiting d3!

fresher-rules: I don't want to have to commit generated rules. The current state management process doesn't trigger rule refreshe considering that state may be committed & saved but rules are not.

streamlined-design: tweak the design rule to be more "freeform" but also considerate of how the design should shape out / what's needed. Include a reference to a tech.md file. 

tests-always: deliver phase isn't reliably testing and creating tests. We should somehow parameterize the rules around testing and all.

feature-complete: ability to complete features and remove them from the active feature list.