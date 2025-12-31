(* ============================================================== *)
(* HEXAGON CLINK: Proving 3 arrangements are insufficient for n=13 *)
(* ============================================================== *)

(*
   PROBLEM: We have 13 items. We want to arrange them on a hexagonal
   grid multiple times, so that every pair of items is adjacent at
   least once across all arrangements.

   QUESTION: Can we do this with just 3 arrangements?

   THE KEY INSIGHT:
   ---------------
   Let's count:
   - Number of pairs we need to cover: C(13,2) = 13*12/2 = 78
   - Number of edges in a 13-node hexagon spiral: 26
   - Total edge-slots in 3 arrangements: 3 * 26 = 78

   The numbers match EXACTLY!

   This means: if 3 arrangements work, every single edge in every
   arrangement must cover a DIFFERENT pair. Zero overlap allowed.

   PROOF STRATEGY:
   --------------
   We show this perfect packing is impossible.

   By symmetry, fix arr1 = {1,2,3,...,13}.
   Then arr2 must share ZERO pairs with arr1.
   Then arr3 must share ZERO pairs with arr1 or arr2.

   We use BACKTRACKING to build arr2 position by position.
   As soon as we place an item that creates an overlapping pair,
   we stop and try a different item. This avoids checking all 13!
   permutations - we prune bad branches early.
*)

(* ------------------------------------------------------------ *)
(* STEP 1: Build the 13-node hexagon spiral                     *)
(* ------------------------------------------------------------ *)

(*
   The "penny spiral" places hexagons one by one:
   - Node 0 at center
   - Each new node must touch the previous node
   - Prefer positions with more neighbors, then closer to center
*)

(* Direction vectors for the 6 neighbors of a hexagon *)
hexDirs = {
  {1.5, 0},        (* Right *)
  {0.75, 1.3},     (* Up-Right *)
  {-0.75, 1.3},    (* Up-Left *)
  {-1.5, 0},       (* Left *)
  {-0.75, -1.3},   (* Down-Left *)
  {0.75, -1.3}     (* Down-Right *)
};

(* Check if two positions are the same (within tolerance) *)
samePosition[p1_, p2_] := Module[{dx, dy},
  dx = Abs[p1[[1]] - p2[[1]]];
  dy = Abs[p1[[2]] - p2[[2]]];
  (dx < 0.1) && (dy < 0.1)
];

(* Check if two positions are adjacent *)
areAdjacent[p1_, p2_] := Module[{diff, isAdj, dir},
  diff = {p2[[1]] - p1[[1]], p2[[2]] - p1[[2]]};
  isAdj = False;
  For[d = 1, d <= 6, d++,
    dir = hexDirs[[d]];
    If[Abs[diff[[1]] - dir[[1]]] < 0.1 && Abs[diff[[2]] - dir[[2]]] < 0.1,
      isAdj = True;
      Break[];
    ];
  ];
  isAdj
];

(* Build the spiral and return the list of edges *)
buildSpiral[n_] := Module[{positions, edges, node, prevPos, cand,
                           bestPos, bestContacts, bestDist, contacts, dist, occupied},

  (* Start with node 0 at the origin *)
  positions = {{0, 0}};
  edges = {};

  (* Add nodes 1 through n-1 *)
  For[node = 1, node <= n - 1, node++,

    (* New node must be adjacent to the previous node *)
    prevPos = positions[[node]];

    (* Find the best position among the 6 neighbors of prevPos *)
    bestPos = Null;
    bestContacts = -1;
    bestDist = 999999;

    For[d = 1, d <= 6, d++,
      cand = prevPos + hexDirs[[d]];

      (* Check if this position is already occupied *)
      occupied = False;
      For[p = 1, p <= Length[positions], p++,
        If[samePosition[positions[[p]], cand],
          occupied = True;
          Break[];
        ];
      ];
      If[occupied, Continue[]];

      (* Count how many existing nodes this position touches *)
      contacts = 0;
      For[p = 1, p <= Length[positions], p++,
        If[areAdjacent[positions[[p]], cand],
          contacts = contacts + 1;
        ];
      ];

      (* Distance from origin (squared) *)
      dist = cand[[1]]^2 + cand[[2]]^2;

      (* Keep best: most contacts, then closest to origin *)
      If[contacts > bestContacts,
        bestPos = cand;
        bestContacts = contacts;
        bestDist = dist;
      ];
      If[contacts == bestContacts && dist < bestDist,
        bestPos = cand;
        bestDist = dist;
      ];
    ];

    (* Add the new node *)
    AppendTo[positions, bestPos];

    (* Add edges to all adjacent existing nodes *)
    For[p = 1, p <= node, p++,
      If[areAdjacent[positions[[p]], bestPos],
        AppendTo[edges, Sort[{p - 1, node}]];
      ];
    ];
  ];

  edges
];

(* ------------------------------------------------------------ *)
(* STEP 2: Set up the problem                                   *)
(* ------------------------------------------------------------ *)

numItems = 13;
hexEdges = buildSpiral[numItems];
numEdges = Length[hexEdges];

Print["============================================"];
Print["HEXAGON STRUCTURE FOR n = ", numItems];
Print["============================================"];
Print[""];
Print["Number of edges: ", numEdges];
Print[""];

(* The crucial arithmetic *)
totalPairs = numItems * (numItems - 1) / 2;
totalSlots = 3 * numEdges;

Print["Pairs to cover:     ", totalPairs];
Print["Slots available:    3 * ", numEdges, " = ", totalSlots];
Print[""];

If[totalSlots < totalPairs,
  Print["IMPOSSIBLE: Not enough edges!"];
  Abort[];
];

If[totalSlots == totalPairs,
  Print["EXACTLY EQUAL: Every edge must cover a unique pair."];
  Print["Zero overlap is allowed between arrangements."];
];

If[totalSlots > totalPairs,
  Print["Slack: ", totalSlots - totalPairs, " edges can overlap."];
];

Print[""];

(* ------------------------------------------------------------ *)
(* STEP 3: Precompute which positions are adjacent              *)
(* ------------------------------------------------------------ *)

(*
   For quick lookup: neighbors[pos] = list of positions adjacent to pos
   Positions are numbered 0 to 12, but Mathematica uses 1-indexing,
   so neighbors[[pos+1]] gives neighbors of position pos.
*)

neighbors = Table[{}, {numItems}];

For[e = 1, e <= numEdges, e++,
  pos1 = hexEdges[[e, 1]];
  pos2 = hexEdges[[e, 2]];
  (* Add each to the other's neighbor list *)
  (* +1 because Mathematica lists are 1-indexed *)
  AppendTo[neighbors[[pos1 + 1]], pos2];
  AppendTo[neighbors[[pos2 + 1]], pos1];
];

Print["Neighbor lists built."];
Print["Example: position 0 is adjacent to positions ", neighbors[[1]]];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 4: Fix arr1 and compute its pairs                       *)
(* ------------------------------------------------------------ *)

(*
   By symmetry, we can set arr1 = {1, 2, 3, ..., 13}.
   This just names the items by which position they start in.
*)

arr1 = Range[numItems];

Print["USING SYMMETRY: arr1 = {1, 2, 3, ..., 13}"];
Print[""];

(* Compute which item-pairs are adjacent in arr1 *)
(* Store them in a lookup table for fast checking *)
(* pairs1Table[i][j] = True if items i and j are adjacent in arr1 *)

pairs1Table = Table[False, {numItems}, {numItems}];

For[e = 1, e <= numEdges, e++,
  pos1 = hexEdges[[e, 1]];
  pos2 = hexEdges[[e, 2]];
  item1 = arr1[[pos1 + 1]];
  item2 = arr1[[pos2 + 1]];
  pairs1Table[[item1, item2]] = True;
  pairs1Table[[item2, item1]] = True;
];

Print["arr1 covers ", numEdges, " pairs."];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 5: Backtracking search for arr2 with zero overlap       *)
(* ------------------------------------------------------------ *)

(*
   We build arr2 one position at a time.
   arr2[[pos+1]] = which item is at position pos.

   After placing an item at position pos, we check:
   - For each neighbor of pos that we've already filled,
     do we create a pair that's already in arr1?
   - If yes, this placement is invalid. Try next item.
   - If no, continue to next position.

   This is MUCH faster than checking all 13! permutations.
*)

Print["============================================"];
Print["SEARCHING FOR arr2 WITH ZERO OVERLAP"];
Print["============================================"];
Print[""];
Print["Using backtracking: build arr2 position by position."];
Print["Prune immediately when any overlap with arr1 is detected."];
Print[""];

(* arr2 will be built up; arr2[[i]] = item at position i-1 *)
(* used[[i]] = True if item i is already placed in arr2 *)

arr2 = Table[0, {numItems}];
used = Table[False, {numItems}];
validArr2List = {};
nodesExplored = 0;

(* Recursive backtracking function *)
(* pos = which position we're filling (0 to 12) *)
searchArr2[pos_] := Module[{item, neighborPos, neighborItem, hasOverlap},

  nodesExplored++;

  (* If we've filled all positions, we found a valid arr2 *)
  If[pos == numItems,
    AppendTo[validArr2List, arr2];
    Return[];
  ];

  (* Try each unused item at this position *)
  For[item = 1, item <= numItems, item++,

    If[used[[item]], Continue[]];  (* Skip if already used *)

    (* Tentatively place this item *)
    arr2[[pos + 1]] = item;
    used[[item]] = True;

    (* Check if this creates any overlap with arr1 *)
    hasOverlap = False;

    (* Look at all neighbors of this position that are already filled *)
    For[k = 1, k <= Length[neighbors[[pos + 1]]], k++,
      neighborPos = neighbors[[pos + 1, k]];

      (* Only check neighbors we've already filled (positions < pos) *)
      If[neighborPos < pos,
        neighborItem = arr2[[neighborPos + 1]];

        (* Is {item, neighborItem} a pair in arr1? *)
        If[pairs1Table[[item, neighborItem]],
          hasOverlap = True;
          Break[];
        ];
      ];
    ];

    (* If no overlap, recurse to fill the next position *)
    If[Not[hasOverlap],
      searchArr2[pos + 1];
    ];

    (* Undo the placement (backtrack) *)
    arr2[[pos + 1]] = 0;
    used[[item]] = False;
  ];
];

(* Run the search *)
startTime = AbsoluteTime[];
searchArr2[0];
endTime = AbsoluteTime[];

Print["Search complete."];
Print["Time: ", Round[endTime - startTime, 0.1], " seconds"];
Print["Nodes explored: ", nodesExplored];
Print["Valid arr2 found: ", Length[validArr2List]];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 6: Check results                                        *)
(* ------------------------------------------------------------ *)

If[Length[validArr2List] == 0,
  Print["============================================"];
  Print["RESULT"];
  Print["============================================"];
  Print[""];
  Print["No arr2 exists with zero overlap with arr1!"];
  Print[""];
  Print["Since we need exactly 78 pairs in 78 edge-slots,"];
  Print["and no arr2 avoids all of arr1's pairs,"];
  Print["a perfect 3-arrangement packing is IMPOSSIBLE."];
  Print[""];
  Print["CONCLUSION: n = 13 requires at least 4 arrangements."];
  Abort[];
];

(* If we found some valid arr2, we need to check arr3 *)
Print["Found ", Length[validArr2List], " valid arr2 candidates."];
Print["Now checking if any allows a valid arr3..."];
Print[""];

(* ------------------------------------------------------------ *)
(* STEP 7: For each valid arr2, search for valid arr3           *)
(* ------------------------------------------------------------ *)

(*
   For each arr2, compute the pairs covered by arr1 and arr2.
   Then search for arr3 that covers exactly the remaining pairs.
*)

(* Helper: get all pairs covered by an arrangement *)
getPairs[arrangement_] := Module[{pairs, pos1, pos2, item1, item2},
  pairs = {};
  For[e = 1, e <= numEdges, e++,
    pos1 = hexEdges[[e, 1]];
    pos2 = hexEdges[[e, 2]];
    item1 = arrangement[[pos1 + 1]];
    item2 = arrangement[[pos2 + 1]];
    AppendTo[pairs, Sort[{item1, item2}]];
  ];
  pairs
];

(* Build table of all 78 pairs *)
allPairs = {};
For[i = 1, i <= numItems - 1, i++,
  For[j = i + 1, j <= numItems, j++,
    AppendTo[allPairs, {i, j}];
  ];
];

foundSolution = False;

For[idx = 1, idx <= Length[validArr2List], idx++,
  currentArr2 = validArr2List[[idx]];
  pairs2 = getPairs[currentArr2];

  (* Pairs still needed = all pairs minus arr1's pairs minus arr2's pairs *)
  (* Build lookup tables *)
  coveredTable = Table[False, {numItems}, {numItems}];

  (* Mark arr1's pairs *)
  For[e = 1, e <= numEdges, e++,
    pos1 = hexEdges[[e, 1]];
    pos2 = hexEdges[[e, 2]];
    item1 = arr1[[pos1 + 1]];
    item2 = arr1[[pos2 + 1]];
    coveredTable[[item1, item2]] = True;
    coveredTable[[item2, item1]] = True;
  ];

  (* Mark arr2's pairs *)
  For[p = 1, p <= Length[pairs2], p++,
    item1 = pairs2[[p, 1]];
    item2 = pairs2[[p, 2]];
    coveredTable[[item1, item2]] = True;
    coveredTable[[item2, item1]] = True;
  ];

  (* Count uncovered pairs *)
  stillNeeded = {};
  For[p = 1, p <= Length[allPairs], p++,
    item1 = allPairs[[p, 1]];
    item2 = allPairs[[p, 2]];
    If[Not[coveredTable[[item1, item2]]],
      AppendTo[stillNeeded, {item1, item2}];
    ];
  ];

  (* arr3 must cover exactly these pairs *)
  (* Use backtracking similar to arr2, but check against stillNeeded *)

  (* Build required pairs table *)
  neededTable = Table[False, {numItems}, {numItems}];
  For[p = 1, p <= Length[stillNeeded], p++,
    item1 = stillNeeded[[p, 1]];
    item2 = stillNeeded[[p, 2]];
    neededTable[[item1, item2]] = True;
    neededTable[[item2, item1]] = True;
  ];

  (* Search for arr3 *)
  arr3 = Table[0, {numItems}];
  used3 = Table[False, {numItems}];
  foundArr3 = False;
  pairsCovered = 0;

  (* For arr3, every edge must be a needed pair (no waste allowed) *)
  searchArr3[pos3_] := Module[{item, nPos, nItem, thisPair, isNeeded, isWaste},

    If[foundArr3, Return[]];

    If[pos3 == numItems,
      (* Check if we covered exactly the right pairs *)
      If[pairsCovered == Length[stillNeeded],
        foundArr3 = True;
      ];
      Return[];
    ];

    For[item = 1, item <= numItems, item++,
      If[used3[[item]], Continue[]];

      arr3[[pos3 + 1]] = item;
      used3[[item]] = True;

      (* Check edges to already-filled neighbors *)
      isWaste = False;
      newPairs = 0;

      For[k = 1, k <= Length[neighbors[[pos3 + 1]]], k++,
        nPos = neighbors[[pos3 + 1, k]];
        If[nPos < pos3,
          nItem = arr3[[nPos + 1]];

          (* Is this pair needed? *)
          If[neededTable[[item, nItem]],
            newPairs++;
          ,
            (* This edge is wasted - not a needed pair *)
            isWaste = True;
            Break[];
          ];
        ];
      ];

      If[Not[isWaste],
        pairsCovered = pairsCovered + newPairs;
        searchArr3[pos3 + 1];
        pairsCovered = pairsCovered - newPairs;
      ];

      arr3[[pos3 + 1]] = 0;
      used3[[item]] = False;
    ];
  ];

  searchArr3[0];

  If[foundArr3,
    Print["*** FOUND A SOLUTION! ***"];
    Print["arr1 = ", arr1];
    Print["arr2 = ", currentArr2];
    Print["arr3 = ", arr3];
    foundSolution = True;
    Break[];
  ];

  If[Mod[idx, 100] == 0,
    Print["Checked ", idx, " / ", Length[validArr2List], " arr2 candidates..."];
  ];
];

(* ------------------------------------------------------------ *)
(* STEP 8: Final result                                         *)
(* ------------------------------------------------------------ *)

Print[""];
Print["============================================"];
Print["FINAL RESULT"];
Print["============================================"];
Print[""];

If[foundSolution,
  Print["3 arrangements ARE sufficient for n = 13."];
,
  Print["3 arrangements are NOT sufficient for n = 13."];
  Print[""];
  Print["We found ", Length[validArr2List], " arr2 candidates with zero"];
  Print["overlap with arr1, but none of them allows an arr3"];
  Print["that covers the remaining pairs without waste."];
  Print[""];
  Print["CONCLUSION: n = 13 requires at least 4 arrangements."];
];
