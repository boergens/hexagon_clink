(* ============================================================== *)
(* HEXAGON CLINK: Proving 3 arrangements are insufficient for n=13 *)
(* v3: Fixed performance - functions defined at top level          *)
(* ============================================================== *)

(* Graph definitions *)
graph1Edges = {
  {2, 3}, {1, 4}, {0, 6}, {5, 6}, {3, 7}, {5, 7},
  {2, 8}, {4, 8}, {0, 9}, {1, 9}, {6, 9}, {1, 10},
  {4, 10}, {8, 10}, {9, 10}, {5, 11}, {6, 11}, {7, 11},
  {9, 11}, {10, 11}, {2, 12}, {3, 12}, {7, 12}, {8, 12},
  {10, 12}, {11, 12}
};

graph2Edges = {
  {1, 3}, {2, 4}, {0, 5}, {0, 6}, {5, 6}, {3, 7},
  {6, 7}, {4, 8}, {5, 8}, {1, 9}, {2, 9}, {5, 10},
  {6, 10}, {7, 10}, {8, 10}, {1, 11}, {3, 11}, {7, 11},
  {9, 11}, {10, 11}, {2, 12}, {4, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

graph3Edges = {
  {0, 3}, {1, 4}, {3, 5}, {4, 6}, {2, 7}, {5, 7},
  {2, 8}, {6, 8}, {0, 9}, {1, 9}, {0, 10}, {3, 10},
  {5, 10}, {7, 10}, {9, 10}, {1, 11}, {4, 11}, {6, 11},
  {8, 11}, {9, 11}, {2, 12}, {7, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

graph4Edges = {
  {0, 2}, {1, 3}, {1, 4}, {0, 5}, {2, 6}, {4, 7},
  {3, 8}, {6, 8}, {5, 9}, {7, 9}, {0, 10}, {2, 10},
  {5, 10}, {6, 10}, {9, 10}, {1, 11}, {3, 11}, {4, 11},
  {7, 11}, {8, 11}, {6, 12}, {7, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

allGraphs = {graph1Edges, graph2Edges, graph3Edges, graph4Edges};
graphNames = {"A", "B", "C", "D"};
numItems = 13;
numEdges = 26;

(* Precompute neighbor lists *)
buildNeighbors[edges_] := Module[{neighbors, pos1, pos2},
  neighbors = Table[{}, {numItems}];
  For[e = 1, e <= Length[edges], e++,
    pos1 = edges[[e, 1]];
    pos2 = edges[[e, 2]];
    AppendTo[neighbors[[pos1 + 1]], pos2];
    AppendTo[neighbors[[pos2 + 1]], pos1];
  ];
  neighbors
];

allNeighbors = Map[buildNeighbors, allGraphs];

(* Build pairs table from arrangement *)
buildPairsTable[edges_, arrangement_] := Module[{table, pos1, pos2, item1, item2},
  table = Table[False, {numItems}, {numItems}];
  For[e = 1, e <= Length[edges], e++,
    pos1 = edges[[e, 1]];
    pos2 = edges[[e, 2]];
    item1 = arrangement[[pos1 + 1]];
    item2 = arrangement[[pos2 + 1]];
    table[[item1, item2]] = True;
    table[[item2, item1]] = True;
  ];
  table
];

allPairs = Flatten[Table[{i, j}, {i, 1, numItems - 1}, {j, i + 1, numItems}], 1];
identityPerm = Range[numItems];

Print["============================================"];
Print["n=13 PROOF (v3 - optimized functions)"];
Print["============================================"];
Print[""];

(* ------------------------------------------------------------ *)
(* Global state for search                                       *)
(* ------------------------------------------------------------ *)

(* These will be set before each search *)
gNeighbors2 = {};
gNeededTable = {};
gArr2 = {};
gUsed2 = {};
gFoundArr2 = False;
gPairsCovered = 0;

(* arr2 search - uses global state *)
searchArr2[pos2_] := Module[{item2, nPos2, nItem2, isWaste, newPairs},
  If[gFoundArr2, Return[]];

  If[pos2 == numItems,
    If[gPairsCovered == numEdges, gFoundArr2 = True];
    Return[];
  ];

  For[item2 = 1, item2 <= numItems, item2++,
    If[gUsed2[[item2]], Continue[]];

    gArr2[[pos2 + 1]] = item2;
    gUsed2[[item2]] = True;

    isWaste = False;
    newPairs = 0;

    For[k = 1, k <= Length[gNeighbors2[[pos2 + 1]]], k++,
      nPos2 = gNeighbors2[[pos2 + 1, k]];
      If[nPos2 < pos2,
        nItem2 = gArr2[[nPos2 + 1]];
        If[gNeededTable[[item2, nItem2]],
          newPairs++;
        ,
          isWaste = True;
          Break[];
        ];
      ];
    ];

    If[Not[isWaste],
      gPairsCovered = gPairsCovered + newPairs;
      searchArr2[pos2 + 1];
      gPairsCovered = gPairsCovered - newPairs;
    ];

    gArr2[[pos2 + 1]] = 0;
    gUsed2[[item2]] = False;
  ];
];

(* Run arr2 search with given parameters *)
runArr2Search[shapeIdx_, neededTable_] := Module[{},
  gNeighbors2 = allNeighbors[[shapeIdx]];
  gNeededTable = neededTable;
  gArr2 = Table[0, {numItems}];
  gUsed2 = Table[False, {numItems}];
  gFoundArr2 = False;
  gPairsCovered = 0;
  searchArr2[0];
  {gFoundArr2, gArr2}
];

(* ------------------------------------------------------------ *)
(* Global state for arr1 search                                  *)
(* ------------------------------------------------------------ *)

gNeighbors1 = {};
gPairs0Table = {};
gArr1 = {};
gUsed1 = {};
gShape1 = 0;
gFoundSolution = False;
gSolutionArr1 = {};
gSolutionArr2 = {};
gSolutionShape2 = 0;
gArr1Count = 0;
gNodesExplored = 0;
gStartTime = 0;

(* arr1 search - uses global state *)
searchArr1[pos1_] := Module[{item, nPos, nItem, hasOverlap, currentArr1,
                              pairs1Table, coveredTable, stillNeeded, neededTable,
                              shape2, result},

  If[gFoundSolution, Return[]];

  gNodesExplored++;
  If[Mod[gNodesExplored, 100000] == 0,
    Print["  nodes: ", gNodesExplored, ", arr1: ", gArr1Count,
          ", time: ", Round[AbsoluteTime[] - gStartTime, 0.1], "s"];
  ];

  (* Complete arr1 found *)
  If[pos1 == numItems,
    gArr1Count++;
    currentArr1 = Table[gArr1[[k]], {k, 1, numItems}];

    If[Mod[gArr1Count, 1000] == 0,
      Print["  ", gArr1Count, " arr1 found..."];
    ];

    pairs1Table = buildPairsTable[allGraphs[[gShape1]], currentArr1];

    (* Compute remaining pairs *)
    stillNeeded = {};
    For[p = 1, p <= Length[allPairs], p++,
      If[Not[gPairs0Table[[allPairs[[p, 1]], allPairs[[p, 2]]]]] &&
         Not[pairs1Table[[allPairs[[p, 1]], allPairs[[p, 2]]]]],
        AppendTo[stillNeeded, allPairs[[p]]];
      ];
    ];

    If[Length[stillNeeded] != numEdges, Return[]];

    neededTable = Table[False, {numItems}, {numItems}];
    For[p = 1, p <= Length[stillNeeded], p++,
      neededTable[[stillNeeded[[p, 1]], stillNeeded[[p, 2]]]] = True;
      neededTable[[stillNeeded[[p, 2]], stillNeeded[[p, 1]]]] = True;
    ];

    (* Try each shape2 >= shape1 *)
    For[shape2 = gShape1, shape2 <= 4, shape2++,
      If[gFoundSolution, Break[]];
      result = runArr2Search[shape2, neededTable];
      If[result[[1]],
        gFoundSolution = True;
        gSolutionArr1 = currentArr1;
        gSolutionArr2 = result[[2]];
        gSolutionShape2 = shape2;
        Return[];
      ];
    ];
    Return[];
  ];

  (* Build arr1 position by position *)
  For[item = 1, item <= numItems, item++,
    If[gFoundSolution, Break[]];
    If[gUsed1[[item]], Continue[]];

    gArr1[[pos1 + 1]] = item;
    gUsed1[[item]] = True;

    hasOverlap = False;
    For[k = 1, k <= Length[gNeighbors1[[pos1 + 1]]], k++,
      nPos = gNeighbors1[[pos1 + 1, k]];
      If[nPos < pos1,
        nItem = gArr1[[nPos + 1]];
        If[gPairs0Table[[item, nItem]],
          hasOverlap = True;
          Break[];
        ];
      ];
    ];

    If[Not[hasOverlap],
      searchArr1[pos1 + 1];
    ];

    gArr1[[pos1 + 1]] = 0;
    gUsed1[[item]] = False;
  ];
];

(* Run arr1 search with given parameters *)
runArr1Search[shape0_, shape1_] := Module[{},
  gNeighbors1 = allNeighbors[[shape1]];
  gPairs0Table = buildPairsTable[allGraphs[[shape0]], identityPerm];
  gArr1 = Table[0, {numItems}];
  gUsed1 = Table[False, {numItems}];
  gShape1 = shape1;
  gArr1Count = 0;
  gNodesExplored = 0;
  gStartTime = AbsoluteTime[];
  searchArr1[0];
  gArr1Count
];

(* ------------------------------------------------------------ *)
(* Main search loop                                              *)
(* ------------------------------------------------------------ *)

globalStartTime = AbsoluteTime[];
gFoundSolution = False;
totalArr1 = 0;

For[shape0 = 1, shape0 <= 4, shape0++,
  If[gFoundSolution, Break[]];

  For[shape1 = shape0, shape1 <= 4, shape1++,
    If[gFoundSolution, Break[]];

    pairName = graphNames[[shape0]] <> graphNames[[shape1]] <> "*";
    Print["Searching ", pairName, " ..."];

    count = runArr1Search[shape0, shape1];
    totalArr1 = totalArr1 + count;

    elapsed = Round[AbsoluteTime[] - gStartTime, 0.1];
    Print["  ", pairName, " done: ", count, " arr1, ", elapsed, "s"];
    Print[""];
  ];
];

(* ------------------------------------------------------------ *)
(* Result                                                        *)
(* ------------------------------------------------------------ *)

totalTime = Round[AbsoluteTime[] - globalStartTime, 0.1];

Print["============================================"];
Print["RESULT"];
Print["============================================"];
Print[""];
Print["Total arr1 found: ", totalArr1];
Print["Total time: ", totalTime, "s"];
Print[""];

If[gFoundSolution,
  Print["*** FOUND A SOLUTION! ***"];
  Print["Shapes: ", graphNames[[shape0]], graphNames[[shape1]], graphNames[[gSolutionShape2]]];
  Print["arr0 = ", identityPerm];
  Print["arr1 = ", gSolutionArr1];
  Print["arr2 = ", gSolutionArr2];
,
  Print["3 arrangements are NOT sufficient for n=13."];
  Print["CONCLUSION: n=13 requires at least 4 arrangements."];
];
