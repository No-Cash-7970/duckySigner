workspace extends workspace.json {
    !script groovy {
        new com.structurizr.autolayout.graphviz.GraphvizAutomaticLayout().apply(workspace);
    }
}
