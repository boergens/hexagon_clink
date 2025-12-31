(* ============================================================== *)
(* HEXAGON CLINK: Proving 3 arrangements are insufficient for n=13 *)
(* OPTIMIZED: symmetry + proper loop nesting                       *)
(* ============================================================== *)

(*
   Loop structure:
   - for shape0 in [1..4]
   - for shape1 in [shape0..4]
   - backtrack to build arr1 (checking legality incrementally)
   - for each valid arr1:
     - for shape2 in [shape1..4]
     - backtrack to build arr2

   This ensures arr1 is generated once per (shape0, shape1), not
   regenerated for each shape2.
*)

(* ------------------------------------------------------------ *)
(* Graph definitions                                             *)
(* ------------------------------------------------------------ *)

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

(* Build neighbor lists *)
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

Print["============================================"];
Print["n=13 PROOF"];
Print["============================================"];
Print[""];
Print["Symmetry: shape0 <= shape1 <= shape2"];
Print["Loop nesting: arr1 search outside shape2 loop"];
Print[""];

(* ------------------------------------------------------------ *)
(* Main search                                                   *)
(* ------------------------------------------------------------ *)

identityPerm = Range[numItems];
allPairs = Flatten[Table[{i, j}, {i, 1, numItems - 1}, {j, i + 1, numItems}], 1];

foundSolution = False;
totalArr1 = 0;
totalArr2Searches = 0;
globalStartTime = AbsoluteTime[];

(* Outer loop: shape0 *)
For[shape0 = 1, shape0 <= 4, shape0++,
  If[foundSolution, Break[]];

  graph0 = allGraphs[[shape0]];
  pairs0Table = buildPairsTable[graph0, identityPerm];

  (* Loop: shape1 >= shape0 *)
  For[shape1 = shape0, shape1 <= 4, shape1++,
    If[foundSolution, Break[]];

    graph1 = allGraphs[[shape1]];
    neighbors1 = allNeighbors[[shape1]];

    pairName = graphNames[[shape0]] <> graphNames[[shape1]] <> "*";
    Print["Searching ", pairName, " ..."];
    pairStartTime = AbsoluteTime[];
    arr1CountThisPair = 0;
    nodesExplored = 0;

    (* Backtracking search for arr1 *)
    arr1 = Table[0, {numItems}];
    used1 = Table[False, {numItems}];

    searchArr1[pos1_] := Module[{item, nPos, nItem, hasOverlap,
                                  currentArr1, pairs1Table, coveredTable,
                                  stillNeeded, neededTable},

      If[foundSolution, Return[]];

      nodesExplored++;
      If[Mod[nodesExplored, 100000] == 0,
        Print["  nodes: ", nodesExplored, ", arr1 found: ", arr1CountThisPair,
              ", time: ", Round[AbsoluteTime[] - pairStartTime, 0.1], "s"];
      ];

      (* Found complete arr1 - now try all shape2 >= shape1 *)
      If[pos1 == numItems,
        totalArr1++;
        arr1CountThisPair++;
        currentArr1 = Table[arr1[[k]], {k, 1, numItems}];
        pairs1Table = buildPairsTable[graph1, currentArr1];

        (* Progress *)
        If[Mod[arr1CountThisPair, 5000] == 0,
          Print["  ", arr1CountThisPair, " arr1 found so far..."];
        ];

        (* Compute covered pairs and remaining *)
        coveredTable = Table[
          pairs0Table[[i, j]] || pairs1Table[[i, j]],
          {i, numItems}, {j, numItems}
        ];

        stillNeeded = {};
        For[p = 1, p <= Length[allPairs], p++,
          If[Not[coveredTable[[allPairs[[p, 1]], allPairs[[p, 2]]]]],
            AppendTo[stillNeeded, allPairs[[p]]];
          ];
        ];

        (* Must have exactly 26 remaining pairs *)
        If[Length[stillNeeded] != numEdges, Return[]];

        neededTable = Table[False, {numItems}, {numItems}];
        For[p = 1, p <= Length[stillNeeded], p++,
          neededTable[[stillNeeded[[p, 1]], stillNeeded[[p, 2]]]] = True;
          neededTable[[stillNeeded[[p, 2]], stillNeeded[[p, 1]]]] = True;
        ];

        (* Inner loop: shape2 >= shape1 *)
        For[shape2 = shape1, shape2 <= 4, shape2++,
          If[foundSolution, Break[]];
          totalArr2Searches++;

          graph2 = allGraphs[[shape2]];
          neighbors2 = allNeighbors[[shape2]];

          (* Backtracking search for arr2 *)
          arr2 = Table[0, {numItems}];
          used2 = Table[False, {numItems}];
          foundArr2 = False;
          pairsCovered = 0;

          searchArr2[pos2_] := Module[{item2, nPos2, nItem2, isWaste, newPairs},
            If[foundArr2, Return[]];

            If[pos2 == numItems,
              If[pairsCovered == numEdges, foundArr2 = True];
              Return[];
            ];

            For[item2 = 1, item2 <= numItems, item2++,
              If[used2[[item2]], Continue[]];

              arr2[[pos2 + 1]] = item2;
              used2[[item2]] = True;

              isWaste = False;
              newPairs = 0;

              For[k = 1, k <= Length[neighbors2[[pos2 + 1]]], k++,
                nPos2 = neighbors2[[pos2 + 1, k]];
                If[nPos2 < pos2,
                  nItem2 = arr2[[nPos2 + 1]];
                  If[neededTable[[item2, nItem2]],
                    newPairs++;
                  ,
                    isWaste = True;
                    Break[];
                  ];
                ];
              ];

              If[Not[isWaste],
                pairsCovered = pairsCovered + newPairs;
                searchArr2[pos2 + 1];
                pairsCovered = pairsCovered - newPairs;
              ];

              arr2[[pos2 + 1]] = 0;
              used2[[item2]] = False;
            ];
          ];

          searchArr2[0];

          If[foundArr2,
            Print[""];
            Print["*** FOUND SOLUTION! ***"];
            Print["Shapes: ", graphNames[[shape0]], graphNames[[shape1]], graphNames[[shape2]]];
            Print["arr0 = ", identityPerm];
            Print["arr1 = ", currentArr1];
            Print["arr2 = ", arr2];
            foundSolution = True;
          ];
        ]; (* end shape2 loop *)

        Return[];
      ];

      (* Build arr1 position by position *)
      For[item = 1, item <= numItems, item++,
        If[foundSolution, Break[]];
        If[used1[[item]], Continue[]];

        arr1[[pos1 + 1]] = item;
        used1[[item]] = True;

        (* Check overlap with arr0 *)
        hasOverlap = False;
        For[k = 1, k <= Length[neighbors1[[pos1 + 1]]], k++,
          nPos = neighbors1[[pos1 + 1, k]];
          If[nPos < pos1,
            nItem = arr1[[nPos + 1]];
            If[pairs0Table[[item, nItem]],
              hasOverlap = True;
              Break[];
            ];
          ];
        ];

        If[Not[hasOverlap],
          searchArr1[pos1 + 1];
        ];

        arr1[[pos1 + 1]] = 0;
        used1[[item]] = False;
      ];
    ];

    searchArr1[0];

    elapsed = Round[AbsoluteTime[] - pairStartTime, 0.1];
    Print["  ", pairName, " done: ", arr1CountThisPair, " arr1, ", elapsed, "s"];
    Print[""];
  ]; (* end shape1 loop *)
]; (* end shape0 loop *)

(* ------------------------------------------------------------ *)
(* Result                                                        *)
(* ------------------------------------------------------------ *)

totalTime = Round[AbsoluteTime[] - globalStartTime, 0.1];

Print["============================================"];
Print["RESULT"];
Print["============================================"];
Print[""];
Print["Total arr1 found: ", totalArr1];
Print["Total arr2 searches: ", totalArr2Searches];
Print["Total time: ", totalTime, "s"];
Print[""];

If[foundSolution,
  Print["3 arrangements ARE sufficient for n=13."];
,
  Print["3 arrangements are NOT sufficient for n=13."];
  Print["CONCLUSION: n=13 requires at least 4 arrangements."];
];
