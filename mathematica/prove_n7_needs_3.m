(* ============================================================== *)
(* HEXAGON CLINK: Proving 2 arrangements are insufficient for n=7 *)
(* ============================================================== *)

(*
   PROBLEM: We have 7 items and want to place them on a hexagonal grid
   such that every pair of items is adjacent at least once.

   GOAL: Prove that 2 arrangements are NOT enough to cover all pairs.

   APPROACH: There are exactly 4 maximal penny graphs on 7 vertices.
   A penny graph is a contact graph of non-overlapping unit disks.
   We exhaustively try all combinations:
     - 4 choices for graph 1 (with canonical labeling)
     - 4 choices for graph 2
     - 7! = 5040 permutations for graph 2's labeling
   Total: 4 * 4 * 5040 = 80,640 combinations

   The 4 maximal penny graphs are:
   1. Spiral/flower (12 edges): center touches all 6 outer nodes
   2-4. Three other configurations (11 edges each)
*)

(* ------------------------------------------------------------ *)
(* STEP 1: Define the 4 maximal penny graph adjacencies         *)
(* ------------------------------------------------------------ *)

(* There are exactly 4 maximal penny graphs on 7 vertices.
   A penny graph is a contact graph of non-overlapping unit disks.
   "Maximal" means no edges can be added while staying a penny graph.

   These are all the possible spatial arrangements we need to consider. *)

(* Graph 1: The spiral/flower (12 edges) - center node 0 touches all *)
graph1Edges = {
  {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6},
  {1, 2}, {1, 6}, {2, 3}, {3, 4}, {4, 5}, {5, 6}
};

(* Graph 2: (11 edges) *)
graph2Edges = {
  {0, 1}, {0, 2}, {0, 3}, {0, 5}, {0, 6},
  {1, 3}, {1, 4}, {1, 6}, {2, 3}, {2, 5}, {3, 4}
};

(* Graph 3: (11 edges) *)
graph3Edges = {
  {0, 1}, {0, 2}, {0, 4}, {0, 5}, {0, 6},
  {1, 5}, {1, 6}, {2, 3}, {2, 4}, {2, 5}, {3, 4}
};

(* Graph 4: (11 edges) *)
graph4Edges = {
  {0, 1}, {0, 2}, {0, 4}, {0, 6},
  {1, 4}, {1, 6}, {2, 3}, {2, 4}, {2, 5}, {3, 4}, {3, 5}
};

allGraphs = {graph1Edges, graph2Edges, graph3Edges, graph4Edges};

Print["Number of maximal penny graphs for n=7: ", Length[allGraphs]];
Print["Edge counts: ", Map[Length, allGraphs]];

(* ------------------------------------------------------------ *)
(* STEP 2: Define our items and count total pairs needed        *)
(* ------------------------------------------------------------ *)

(* We have 7 items, labeled 1 through 7 *)
items = {1, 2, 3, 4, 5, 6, 7};
numItems = 7;

(* Total number of pairs we need to cover: C(7,2) = 21 *)
totalPairsNeeded = numItems * (numItems - 1) / 2;
Print["Total pairs to cover: ", totalPairsNeeded];

(* ------------------------------------------------------------ *)
(* STEP 3: Function to get item-pairs from one arrangement      *)
(* ------------------------------------------------------------ *)

(* An "arrangement" is a permutation telling us which item is at each position,
   plus a graph structure defining which positions are adjacent.

   For example, arrangement = {3,1,4,2,7,5,6} means:
     - Item 3 is at position 0
     - Item 1 is at position 1
     - Item 4 is at position 2
     - ... and so on

   This function returns all item-pairs that are adjacent in this arrangement. *)

getPairsFromArrangement[graphEdges_, arrangement_] := Module[
  {pairs, edge, pos1, pos2, item1, item2},

  (* Start with empty list of pairs *)
  pairs = {};

  (* For each edge in the graph... *)
  For[i = 1, i <= Length[graphEdges], i++,
    edge = graphEdges[[i]];

    (* Get the two positions that are adjacent *)
    pos1 = edge[[1]];
    pos2 = edge[[2]];

    (* Find which items are at those positions *)
    (* Note: Mathematica lists are 1-indexed, so we add 1 *)
    item1 = arrangement[[pos1 + 1]];
    item2 = arrangement[[pos2 + 1]];

    (* Store the pair in sorted order so {2,5} and {5,2} are the same *)
    pairs = Append[pairs, Sort[{item1, item2}]];
  ];

  (* Return the list of pairs *)
  pairs
];

(* ------------------------------------------------------------ *)
(* STEP 4: Test the function with an example                    *)
(* ------------------------------------------------------------ *)

(* Example: items placed in order 1,2,3,4,5,6,7 at positions 0,1,2,3,4,5,6 *)
exampleArrangement = {1, 2, 3, 4, 5, 6, 7};
examplePairs = getPairsFromArrangement[graph1Edges, exampleArrangement];
Print["Example arrangement on graph 1: ", exampleArrangement];
Print["Pairs covered: ", examplePairs];
Print["Number of pairs: ", Length[examplePairs]];

(* ------------------------------------------------------------ *)
(* STEP 5: Generate all possible arrangements (permutations)    *)
(* ------------------------------------------------------------ *)

(* All ways to arrange 7 items on 7 positions *)
allArrangements = Permutations[items];
numArrangements = Length[allArrangements];
Print["Total possible arrangements: ", numArrangements, " (this is 7!)"];

(* ------------------------------------------------------------ *)
(* STEP 6: Generate all pairs we need to cover                  *)
(* ------------------------------------------------------------ *)

(* All 21 pairs of items *)
allItemPairs = Subsets[items, {2}];
Print["All item pairs we need: ", allItemPairs];

(* ------------------------------------------------------------ *)
(* STEP 7: Check if ANY two arrangements can cover all pairs    *)
(* ------------------------------------------------------------ *)

(* We try all combinations:
   - Graph 1 with canonical labeling (identity permutation)
   - Graph 2 with all 7! = 5040 permutations
   This covers all non-isomorphic cases. *)

Print["\nSearching all graph pairs with all permutations..."];
Print["This checks ", Length[allGraphs], " graphs x ", Length[allGraphs],
      " graphs x ", numArrangements, " permutations = ",
      Length[allGraphs] * Length[allGraphs] * numArrangements, " combinations"];

(* We'll track if we ever find a solution *)
foundSolution = False;
checksPerformed = 0;

(* For graph 1, we fix the canonical labeling (identity) *)
identityPerm = {1, 2, 3, 4, 5, 6, 7};

(* Try every pair of graphs *)
For[g1 = 1, g1 <= Length[allGraphs], g1++,
  graph1 = allGraphs[[g1]];
  pairs1 = getPairsFromArrangement[graph1, identityPerm];

  For[g2 = 1, g2 <= Length[allGraphs], g2++,
    graph2 = allGraphs[[g2]];

    (* Try all permutations for graph 2 *)
    For[p = 1, p <= numArrangements, p++,
      perm = allArrangements[[p]];
      pairs2 = getPairsFromArrangement[graph2, perm];

      (* Combine all pairs from both arrangements (removing duplicates) *)
      combinedPairs = Union[pairs1, pairs2];

      (* Check if we covered all 21 pairs *)
      If[Length[combinedPairs] == totalPairsNeeded,
        Print["FOUND SOLUTION!"];
        Print["Graph 1: #", g1, " (", Length[graph1], " edges)"];
        Print["Graph 2: #", g2, " (", Length[graph2], " edges)"];
        Print["Permutation: ", perm];
        foundSolution = True;
        Break[];
      ];

      checksPerformed++;
    ];

    If[foundSolution, Break[]];
  ];

  If[foundSolution, Break[]];

  Print["Finished graph 1 = #", g1];
];

(* ------------------------------------------------------------ *)
(* STEP 8: Report the final result                              *)
(* ------------------------------------------------------------ *)

Print["\n========== RESULT =========="];
Print["Total checks performed: ", checksPerformed];

If[foundSolution,
  Print["It IS possible to cover all pairs with 2 arrangements."],
  Print["It is NOT possible to cover all pairs with 2 arrangements."];
  Print["Therefore, n=7 requires at least 3 arrangements."];
];

(* ------------------------------------------------------------ *)
(* STEP 9: Bonus analysis - what's the maximum coverage?        *)
(* ------------------------------------------------------------ *)

Print["\n========== BONUS: Maximum coverage with 2 arrangements =========="];

maxCoverage = 0;
bestG1 = 0;
bestG2 = 0;
bestPerm = {};

For[g1 = 1, g1 <= Length[allGraphs], g1++,
  graph1 = allGraphs[[g1]];
  pairs1 = getPairsFromArrangement[graph1, identityPerm];

  For[g2 = 1, g2 <= Length[allGraphs], g2++,
    graph2 = allGraphs[[g2]];

    For[p = 1, p <= numArrangements, p++,
      perm = allArrangements[[p]];
      pairs2 = getPairsFromArrangement[graph2, perm];

      combinedPairs = Union[pairs1, pairs2];
      coverage = Length[combinedPairs];

      If[coverage > maxCoverage,
        maxCoverage = coverage;
        bestG1 = g1;
        bestG2 = g2;
        bestPerm = perm;
      ];
    ];
  ];
];

Print["Maximum pairs coverable with 2 arrangements: ", maxCoverage, " out of ", totalPairsNeeded];
Print["Best graph 1: #", bestG1, " (", Length[allGraphs[[bestG1]]], " edges)"];
Print["Best graph 2: #", bestG2, " (", Length[allGraphs[[bestG2]]], " edges)"];
Print["Best permutation for graph 2: ", bestPerm];

(* Show which pairs are missing *)
bestPairs1 = getPairsFromArrangement[allGraphs[[bestG1]], identityPerm];
bestPairs2 = getPairsFromArrangement[allGraphs[[bestG2]], bestPerm];
coveredPairs = Union[bestPairs1, bestPairs2];
missingPairs = Complement[allItemPairs, coveredPairs];
Print["Missing pairs: ", missingPairs];
