(* ============================================================== *)
(* HEXAGON CLINK: Proving 3 arrangements are insufficient for n=13 *)
(* COMPLETE VERSION: Tests all 4 maximal penny graphs              *)
(* ============================================================== *)

(*
   PROBLEM: We have 13 items. We want to arrange them on a hexagonal
   grid multiple times, so that every pair of items is adjacent at
   least once across all arrangements.

   QUESTION: Can we do this with just 3 arrangements?

   THE KEY INSIGHT:
   ---------------
   - Number of pairs we need to cover: C(13,2) = 78
   - Number of edges in ANY maximal 13-node penny graph: 26
   - Total edge-slots in 3 arrangements: 3 * 26 = 78

   The numbers match EXACTLY! Zero overlap allowed.

   IMPORTANT: There are exactly 4 non-isomorphic maximal penny graphs
   on 13 vertices (found via polyiamond enumeration). We must test
   ALL combinations of graphs for arr0, arr1, arr2.

   Total graph combinations: 4 * 4 * 4 = 64
*)

(* ------------------------------------------------------------ *)
(* STEP 1: Define all 4 maximal penny graphs on 13 vertices      *)
(* ------------------------------------------------------------ *)

(* These were found via polyiamond enumeration and verified as
   the complete set of non-isomorphic maximal penny graphs on
   13 vertices with 26 edges. *)

(* Graph 1: 13 vertices, 26 edges *)
graph1Edges = {
  {2, 3}, {1, 4}, {0, 6}, {5, 6}, {3, 7}, {5, 7},
  {2, 8}, {4, 8}, {0, 9}, {1, 9}, {6, 9}, {1, 10},
  {4, 10}, {8, 10}, {9, 10}, {5, 11}, {6, 11}, {7, 11},
  {9, 11}, {10, 11}, {2, 12}, {3, 12}, {7, 12}, {8, 12},
  {10, 12}, {11, 12}
};

(* Graph 2: 13 vertices, 26 edges *)
graph2Edges = {
  {1, 3}, {2, 4}, {0, 5}, {0, 6}, {5, 6}, {3, 7},
  {6, 7}, {4, 8}, {5, 8}, {1, 9}, {2, 9}, {5, 10},
  {6, 10}, {7, 10}, {8, 10}, {1, 11}, {3, 11}, {7, 11},
  {9, 11}, {10, 11}, {2, 12}, {4, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

(* Graph 3: 13 vertices, 26 edges *)
graph3Edges = {
  {0, 3}, {1, 4}, {3, 5}, {4, 6}, {2, 7}, {5, 7},
  {2, 8}, {6, 8}, {0, 9}, {1, 9}, {0, 10}, {3, 10},
  {5, 10}, {7, 10}, {9, 10}, {1, 11}, {4, 11}, {6, 11},
  {8, 11}, {9, 11}, {2, 12}, {7, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

(* Graph 4: 13 vertices, 26 edges *)
graph4Edges = {
  {0, 2}, {1, 3}, {1, 4}, {0, 5}, {2, 6}, {4, 7},
  {3, 8}, {6, 8}, {5, 9}, {7, 9}, {0, 10}, {2, 10},
  {5, 10}, {6, 10}, {9, 10}, {1, 11}, {3, 11}, {4, 11},
  {7, 11}, {8, 11}, {6, 12}, {7, 12}, {8, 12}, {9, 12},
  {10, 12}, {11, 12}
};

allGraphs = {graph1Edges, graph2Edges, graph3Edges, graph4Edges};
numGraphs = 4;
numItems = 13;
numEdges = 26;

Print["============================================"];
Print["COMPLETE PROOF: Testing all 4 maximal penny graphs"];
Print["============================================"];
Print[""];
Print["Number of maximal penny graphs on 13 vertices: ", numGraphs];
Print["Each graph has ", numEdges, " edges"];
Print[""];

(* The crucial arithmetic *)
totalPairs = numItems * (numItems - 1) / 2;
totalSlots = 3 * numEdges;

Print["Pairs to cover:     ", totalPairs];
Print["Slots available:    3 * ", numEdges, " = ", totalSlots];
Print[""];
Print["EXACTLY EQUAL: Every edge must cover a unique pair."];
Print["Zero overlap allowed between ANY two arrangements."];
Print[""];
Print["Graph combinations to test: ", numGraphs, "^3 = ", numGraphs^3];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 2: Precompute neighbor lists for each graph             *)
(* ------------------------------------------------------------ *)

(* For each graph, build a neighbor lookup table *)
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
Print["Neighbor lists built for all graphs."];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 3: Build pair lookup table from an arrangement          *)
(* ------------------------------------------------------------ *)

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

(* ------------------------------------------------------------ *)
(* STEP 4: Search for valid arrangements                        *)
(* ------------------------------------------------------------ *)

(* All 78 pairs of items *)
allPairs = Flatten[Table[{i, j}, {i, 1, numItems - 1}, {j, i + 1, numItems}], 1];

(* Identity arrangement *)
identityPerm = Range[numItems];

foundSolution = False;
totalCombinations = 0;
totalArr1Found = 0;

Print["============================================"];
Print["SEARCHING ALL GRAPH COMBINATIONS"];
Print["============================================"];
Print[""];

(* For each graph triple (g0, g1, g2) *)
For[g0 = 1, g0 <= numGraphs, g0++,
  If[foundSolution, Break[]];

  graph0 = allGraphs[[g0]];
  neighbors0 = allNeighbors[[g0]];

  (* Fix arr0 = identity on graph0 *)
  pairs0Table = buildPairsTable[graph0, identityPerm];

  For[g1 = 1, g1 <= numGraphs, g1++,
    If[foundSolution, Break[]];

    graph1 = allGraphs[[g1]];
    neighbors1 = allNeighbors[[g1]];

    (* Search for arr1 with zero overlap with arr0 *)
    arr1 = Table[0, {numItems}];
    used1 = Table[False, {numItems}];
    validArr1List = {};

    searchArr1[pos_] := Module[{item, nPos, nItem, hasOverlap},
      If[pos == numItems,
        AppendTo[validArr1List, arr1];
        Return[];
      ];

      For[item = 1, item <= numItems, item++,
        If[used1[[item]], Continue[]];

        arr1[[pos + 1]] = item;
        used1[[item]] = True;

        hasOverlap = False;
        For[k = 1, k <= Length[neighbors1[[pos + 1]]], k++,
          nPos = neighbors1[[pos + 1, k]];
          If[nPos < pos,
            nItem = arr1[[nPos + 1]];
            If[pairs0Table[[item, nItem]],
              hasOverlap = True;
              Break[];
            ];
          ];
        ];

        If[Not[hasOverlap],
          searchArr1[pos + 1];
        ];

        arr1[[pos + 1]] = 0;
        used1[[item]] = False;
      ];
    ];

    searchArr1[0];

    If[Length[validArr1List] > 0,
      Print["Graphs (", g0, ",", g1, "): found ", Length[validArr1List], " valid arr1"];
      totalArr1Found = totalArr1Found + Length[validArr1List];

      (* For each valid arr1, try each graph2 *)
      For[a1Idx = 1, a1Idx <= Length[validArr1List], a1Idx++,
        If[foundSolution, Break[]];

        currentArr1 = validArr1List[[a1Idx]];
        pairs1Table = buildPairsTable[graph1, currentArr1];

        (* Combine pairs from arr0 and arr1 *)
        coveredTable = Table[False, {numItems}, {numItems}];
        For[i = 1, i <= numItems, i++,
          For[j = 1, j <= numItems, j++,
            coveredTable[[i, j]] = pairs0Table[[i, j]] || pairs1Table[[i, j]];
          ];
        ];

        (* Pairs still needed *)
        stillNeeded = {};
        For[p = 1, p <= Length[allPairs], p++,
          item1 = allPairs[[p, 1]];
          item2 = allPairs[[p, 2]];
          If[Not[coveredTable[[item1, item2]]],
            AppendTo[stillNeeded, {item1, item2}];
          ];
        ];

        (* Try each graph for arr2 *)
        For[g2 = 1, g2 <= numGraphs, g2++,
          If[foundSolution, Break[]];

          totalCombinations++;

          graph2 = allGraphs[[g2]];
          neighbors2 = allNeighbors[[g2]];

          (* arr2 must cover exactly stillNeeded (26 pairs) with no waste *)
          neededTable = Table[False, {numItems}, {numItems}];
          For[p = 1, p <= Length[stillNeeded], p++,
            item1 = stillNeeded[[p, 1]];
            item2 = stillNeeded[[p, 2]];
            neededTable[[item1, item2]] = True;
            neededTable[[item2, item1]] = True;
          ];

          arr2 = Table[0, {numItems}];
          used2 = Table[False, {numItems}];
          foundArr2 = False;
          pairsCovered = 0;

          searchArr2[pos2_] := Module[{item, nPos, nItem, isWaste, newPairs},
            If[foundArr2, Return[]];

            If[pos2 == numItems,
              If[pairsCovered == Length[stillNeeded],
                foundArr2 = True;
              ];
              Return[];
            ];

            For[item = 1, item <= numItems, item++,
              If[used2[[item]], Continue[]];

              arr2[[pos2 + 1]] = item;
              used2[[item]] = True;

              isWaste = False;
              newPairs = 0;

              For[k = 1, k <= Length[neighbors2[[pos2 + 1]]], k++,
                nPos = neighbors2[[pos2 + 1, k]];
                If[nPos < pos2,
                  nItem = arr2[[nPos + 1]];
                  If[neededTable[[item, nItem]],
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
              used2[[item]] = False;
            ];
          ];

          searchArr2[0];

          If[foundArr2,
            Print[""];
            Print["*** FOUND A SOLUTION! ***"];
            Print["Graph triple: (", g0, ", ", g1, ", ", g2, ")"];
            Print["arr0 = ", identityPerm];
            Print["arr1 = ", currentArr1];
            Print["arr2 = ", arr2];
            foundSolution = True;
          ];
        ];
      ];
    ];
  ];

  Print["Completed graph0 = ", g0, " / ", numGraphs];
];

(* ------------------------------------------------------------ *)
(* STEP 5: Final result                                         *)
(* ------------------------------------------------------------ *)

Print[""];
Print["============================================"];
Print["FINAL RESULT"];
Print["============================================"];
Print[""];
Print["Total graph combinations tested: ", numGraphs, "^3 = 64"];
Print["Total valid arr1 arrangements found: ", totalArr1Found];
Print["Total (arr0, arr1, g2) triples checked: ", totalCombinations];
Print[""];

If[foundSolution,
  Print["3 arrangements ARE sufficient for n = 13."];
,
  Print["3 arrangements are NOT sufficient for n = 13."];
  Print[""];
  Print["We tested ALL 4 maximal penny graphs in every position"];
  Print["(arr0, arr1, arr2) and found no valid 3-arrangement."];
  Print[""];
  Print["CONCLUSION: n = 13 requires at least 4 arrangements."];
];
