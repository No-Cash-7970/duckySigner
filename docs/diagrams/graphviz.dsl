workspace extends workspace.json {
    !script groovy {
        new com.structurizr.graphviz.GraphvizAutomaticLayout().apply(workspace);
    }
}