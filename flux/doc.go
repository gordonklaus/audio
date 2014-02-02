/*
Flux is a dataflow-style graphical editor for writing Go programs.


Browser

The browser is the first thing you see when starting Flux.  It provides a means of navigating the directories and packages under GOPATH and in the standard library and the objects within those packages, and of creating, deleting, and selecting such items.

Packages and directories are displayed in white, types in green, functions and methods in red, variables, struct fields, and constants in blue, and special items in yellow.  Use the up and down arrow keys to scroll through the list.  Type a prefix to filter the list.  When a package, directory, or type name is highlighted, press the right arrow key to view its children.  Press the left arrow key to go back to the parent.  Press Enter to select the current item.

To create a new item, hold Command and press 1 (package or directory), 2 (type), 3 (func or method), 4 (var or struct field), or 5 (const); then, type the new item's name followed by Enter.  The new item will be opened for editing.

To delete an item, press Command-Delete (not yet implemented).

To change the name of a package, press Shift-Enter, then edit the name and press Enter.

The browser behaves differently depending on the context in which it is opened.  In the context of program start, it displays only objects created in Flux and it allows you to create, delete, or open them for editing.  When opened in the context of editing a type or function, a relevant subset of objects is displayed from which one can be selected.


Function editor

The function editor displays a function or method as a kind of graph.  The nodes of the graph specify operations such as function calls and control flow.  Nodes typically have some inputs and outputs (generally, ports) by which they can be connected.  A connection has an output as its source and an input as its destination, indicating that a value is passed from the output to the input.  An input may have zero or more connections; the value used is the last one to have been passed or the zero value if none.

Every node belongs to a block.  Outermost is the function block, all of whose nodes are executed when the function is called.  An if-node has an if-block and an else-block, each containing nodes that are conditionally executed.  A loop node has a loop block whose nodes are executed zero or more times.  A function literal node has a function block whose nodes are executed when the function value is called.

The execution order of nodes is determined as follows:  Node A runs before node B if there is a connection with A as its source and B as its destination.  A connection that exits or enters a block has that block's containing node as a source or destination, respectively.

The arrow keys are used to navigate the graph.  On their own, they move the focus along and between connections following the topology of the graph.  While holding Option, they move the focus between nodes with no regard for connectivity.  Pressing Escape moves the focus from a connection end to its port or from a node to its containing node.

To create a named node (function or method, struct field, operator, special node), simply start typing its name; the browser will open, allowing you to select the desired item.  Hold Shift in the browser to treat functions and methods as values; otherwise they are treated as calls.

A struct field node can be toggled between read and write using the Equals key.

A method value node with an unconnected receiver is treated as a method expression.

A variadic function or method call node can have inputs added by pressing the Comma key, and can be toggled between multiple element input and single slice input modes by pressing Control-Period.

To create a numeric, string, or character literal node, type a digit, double quote, or single quote character, respectively, followed by the value and Enter.  To create a composite literal node, type a left curly brace character and select the desired type from the browser.

A function block always has at least two nodes, one for parameters and another for results.  To add a parameter or result, focus the appropriate node or port and press Comma (hold Shift to insert before a port), type the name and Enter, then select the type from the browser.  To delete a parameter or result, focus the port and press Backspace or Delete.  To toggle the signature's variadicity, focus the final parameter's port and press Control-Period.

To create a new connection, move the focus to a port and press Enter to start editing.  Use the arrow keys to move the other end of the connection and press Enter to stop editing.  To edit an existing connection, move the focus to one of its ends and press Enter.

As an alternative to being drawn as a line, a connection may be named by pressing Underscore and typing a name followed by Enter.

Press Backspace or Delete to delete a node or connection.

To save changes, press Command-S.


Type editor

The type editor displays a type as a tree.  Composite types have their children nested inside them; named types are leaves.

Press Enter to move the focus from a composite type to one of its children.  Use the arrow keys to move the focus between the children of a composite type.  Press Escape to move the focus from a child to its parent.

To edit the focused item, press Backspace.  For a named item (struct field, function parameter or result, and interface method), first type the name and Enter.  Otherwise just select the type from the browser.  After a composite type is created, each of its children will be edited.  Press Escape to stop entering new named items.  Press Comma to insert a new named item (hold Shift to insert before the focused item); to delete one, press Delete.
*/
package main
