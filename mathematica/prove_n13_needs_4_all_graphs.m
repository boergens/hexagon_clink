(* ============================================================== *)
(* HEXAGON CLINK: Proving 3 arrangements are insufficient for n=13 *)
(* v4: Simple approach - iterate over 20 combinations, run v1 logic *)
(* ============================================================== *)

(* All 4 maximal penny graphs on 13 vertices *)
allGraphEdges = {
  (* Graph A = SPIRAL (good node ordering: node 0 has degree 6) *)
  {{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {1, 2}, {1, 6}, {1, 7}, {1, 8}, {1, 9}, {2, 3}, {2, 9}, {2, 10}, {2, 11}, {3, 4}, {3, 11}, {3, 12}, {4, 5}, {5, 6}, {6, 7}, {7, 8}, {8, 9}, {9, 10}, {10, 11}, {11, 12}},
  (* Graph B *)
  {{1, 3}, {2, 4}, {0, 5}, {0, 6}, {5, 6}, {3, 7}, {6, 7}, {4, 8}, {5, 8}, {1, 9}, {2, 9}, {5, 10}, {6, 10}, {7, 10}, {8, 10}, {1, 11}, {3, 11}, {7, 11}, {9, 11}, {10, 11}, {2, 12}, {4, 12}, {8, 12}, {9, 12}, {10, 12}, {11, 12}},
  (* Graph C *)
  {{0, 3}, {1, 4}, {3, 5}, {4, 6}, {2, 7}, {5, 7}, {2, 8}, {6, 8}, {0, 9}, {1, 9}, {0, 10}, {3, 10}, {5, 10}, {7, 10}, {9, 10}, {1, 11}, {4, 11}, {6, 11}, {8, 11}, {9, 11}, {2, 12}, {7, 12}, {8, 12}, {9, 12}, {10, 12}, {11, 12}},
  (* Graph D *)
  {{0, 2}, {1, 3}, {1, 4}, {0, 5}, {2, 6}, {4, 7}, {3, 8}, {6, 8}, {5, 9}, {7, 9}, {0, 10}, {2, 10}, {5, 10}, {6, 10}, {9, 10}, {1, 11}, {3, 11}, {4, 11}, {7, 11}, {8, 11}, {6, 12}, {7, 12}, {8, 12}, {9, 12}, {10, 12}, {11, 12}}
};
graphNames = {"A", "B", "C", "D"};

numItems = 13;
numEdges = 26;
totalPairs = numItems * (numItems - 1) / 2;  (* 78 *)

Print["============================================"];
Print["n=13 PROOF (v4 - simple iteration)"];
Print["============================================"];
Print[""];
Print["Testing all 20 shape combinations (shape0 <= shape1 <= shape2)"];
Print[""];

(* Generate 20 combinations *)
combos = {};
For[s0 = 1, s0 <= 4, s0++,
  For[s1 = s0, s1 <= 4, s1++,
    For[s2 = s1, s2 <= 4, s2++,
      AppendTo[combos, {s0, s1, s2}];
    ];
  ];
];

Print["Combinations: ", Length[combos]];
Print[""];

(* Helper: build neighbor list *)
buildNeighbors[edges_] := Module[{neighbors},
  neighbors = Table[{}, {numItems}];
  Do[
    AppendTo[neighbors[[edges[[e, 1]] + 1]], edges[[e, 2]]];
    AppendTo[neighbors[[edges[[e, 2]] + 1]], edges[[e, 1]]];
  , {e, Length[edges]}];
  neighbors
];

(* Helper: get pairs from arrangement *)
getPairs[edges_, arrangement_] := Module[{pairs},
  pairs = Table[False, {numItems}, {numItems}];
  Do[
    item1 = arrangement[[edges[[e, 1]] + 1]];
    item2 = arrangement[[edges[[e, 2]] + 1]];
    pairs[[item1, item2]] = True;
    pairs[[item2, item1]] = True;
  , {e, Length[edges]}];
  pairs
];

globalStart = AbsoluteTime[];
foundSolution = False;
solutionCombo = {};
solutionArr1 = {};
solutionArr2 = {};

(* Main loop over 20 combinations *)
For[ci = 1, ci <= Length[combos], ci++,
  If[foundSolution, Break[]];

  s0 = combos[[ci, 1]];
  s1 = combos[[ci, 2]];
  s2 = combos[[ci, 3]];

  comboName = graphNames[[s0]] <> graphNames[[s1]] <> graphNames[[s2]];
  Print["[", ci, "/20] ", comboName, " ..."];
  comboStart = AbsoluteTime[];

  edges0 = allGraphEdges[[s0]];
  edges1 = allGraphEdges[[s1]];
  edges2 = allGraphEdges[[s2]];
  neighbors1 = buildNeighbors[edges1];
  neighbors2 = buildNeighbors[edges2];

  (* arr0 = identity on shape0 *)
  arr0 = Range[numItems];
  pairs0 = getPairs[edges0, arr0];

  (* Search for arr1 on shape1 with zero overlap *)
  arr1 = Table[0, {numItems}];
  used1 = Table[False, {numItems}];
  validArr1List = {};
  nodesExplored = 0;

  searchArr1[pos_] := Module[{item, nPos, nItem, hasOverlap},
    If[foundSolution, Return[]];

    nodesExplored++;

    If[pos == numItems,
      AppendTo[validArr1List, Table[arr1[[k]], {k, numItems}]];
      Return[];
    ];

    Do[
      If[used1[[item]], Continue[]];

      arr1[[pos + 1]] = item;
      used1[[item]] = True;

      hasOverlap = False;
      Do[
        nPos = neighbors1[[pos + 1, k]];
        If[nPos < pos,
          nItem = arr1[[nPos + 1]];
          If[pairs0[[item, nItem]],
            hasOverlap = True;
            Break[];
          ];
        ];
      , {k, Length[neighbors1[[pos + 1]]]}];

      If[Not[hasOverlap], searchArr1[pos + 1]];

      arr1[[pos + 1]] = 0;
      used1[[item]] = False;
    , {item, numItems}];
  ];

  searchArr1[0];

  Print["  Found ", Length[validArr1List], " arr1 candidates (", nodesExplored, " nodes, ",
        Round[AbsoluteTime[] - comboStart, 0.1], "s)"];

  (* For each arr1, search for arr2 on shape2 *)
  Do[
    If[foundSolution, Break[]];

    currentArr1 = validArr1List[[ai]];
    pairs1 = getPairs[edges1, currentArr1];

    (* Compute needed pairs *)
    neededTable = Table[False, {numItems}, {numItems}];
    neededCount = 0;
    Do[
      If[Not[pairs0[[i, j]]] && Not[pairs1[[i, j]]],
        neededTable[[i, j]] = True;
        neededTable[[j, i]] = True;
        neededCount++;
      ];
    , {i, numItems - 1}, {j, i + 1, numItems}];

    If[neededCount != numEdges, Continue[]];

    (* Search arr2 *)
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

      Do[
        If[used2[[item2]], Continue[]];

        arr2[[pos2 + 1]] = item2;
        used2[[item2]] = True;

        isWaste = False;
        newPairs = 0;

        Do[
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
        , {k, Length[neighbors2[[pos2 + 1]]]}];

        If[Not[isWaste],
          pairsCovered = pairsCovered + newPairs;
          searchArr2[pos2 + 1];
          pairsCovered = pairsCovered - newPairs;
        ];

        arr2[[pos2 + 1]] = 0;
        used2[[item2]] = False;
      , {item2, numItems}];
    ];

    searchArr2[0];

    If[foundArr2,
      foundSolution = True;
      solutionCombo = {s0, s1, s2};
      solutionArr1 = currentArr1;
      solutionArr2 = Table[arr2[[k]], {k, numItems}];
    ];

    If[Mod[ai, 100] == 0,
      Print["    Checked ", ai, "/", Length[validArr1List], " arr1..."];
    ];
  , {ai, Length[validArr1List]}];

  Print["  ", comboName, " done in ", Round[AbsoluteTime[] - comboStart, 0.1], "s"];
  Print[""];
];

(* Result *)
Print["============================================"];
Print["RESULT"];
Print["============================================"];
Print[""];
Print["Total time: ", Round[AbsoluteTime[] - globalStart, 0.1], "s"];
Print[""];

If[foundSolution,
  Print["*** FOUND SOLUTION! ***"];
  Print["Shapes: ", graphNames[[solutionCombo[[1]]]], graphNames[[solutionCombo[[2]]]], graphNames[[solutionCombo[[3]]]]];
  Print["arr0 = ", Range[numItems]];
  Print["arr1 = ", solutionArr1];
  Print["arr2 = ", solutionArr2];
,
  Print["No solution found."];
  Print["3 arrangements are NOT sufficient for n=13."];
  Print["CONCLUSION: n=13 requires at least 4 arrangements."];
];
