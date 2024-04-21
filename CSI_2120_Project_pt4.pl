% ==========================================
% Ben Miller -- 300297574
% Arin Barak -- 300280812
% CSI-2120 Fall 2024 Course Project part 4
% ==========================================


% dataset(DirectoryName)
% this is where the image dataset is located
dataset('C:\\Users\\benem\\Downloads\\Dataset\\imageDataset2_15_20\\').

% directory_textfiles(DirectoryName, ListOfTextfiles)
% produces the list of text files in a directory
directory_textfiles(D,Textfiles):- directory_files(D,Files), include(isTextFile, Files, Textfiles).

isTextFile(Filename):-string_concat(_,'.txt',Filename).


% read_hist_file(Filename,ListOfNumbers)
% reads a histogram file, normalizes the histogram, and produces a list of numbers (bin values)
read_hist_file(Filename,NormalizedNumbers):-
    open(Filename,read,Stream),
    read_line_to_string(Stream,_),
    read_line_to_string(Stream,String),
    close(Stream),
    atomic_list_concat(List, ' ', String),
    atoms_numbers(List,Numbers),
    normalize_histogram(Numbers, NormalizedNumbers).  % added to make sure histograms relatedness is in [0.0 , 1.0]



% similarity_search(QueryFile,SimilarImageList)
% returns the list of images similar to the query image
% similar images are specified as (ImageName, SimilarityScore)
% predicat dataset/1 provides the location of the image set
similarity_search(QueryFile,SimilarList) :- dataset(D), directory_textfiles(D,TxtFiles),
                                            similarity_search(QueryFile,D,TxtFiles,SimilarList).


% similarity_search(QueryFile, DatasetDirectory, HistoFileList, SimilarImageList)
similarity_search(QueryFile,DatasetDirectory, DatasetFiles,Best):- read_hist_file(QueryFile,QueryHisto),
                                            compare_histograms(QueryHisto, DatasetFiles, DatasetDirectory, Scores),
                                            sort(2,@>,Scores,Sorted),take(Sorted,5,Best).



% compare_histograms(QueryHisto, DatasetFiles, DatasetDirectory, FileScores)
% compares a query histogram with a list of histogram files
compare_histograms(QueryHisto, DatasetFiles, DatasetDirectory, FileScores) :-
    compare_histograms_helper(QueryHisto, DatasetFiles, DatasetDirectory, [], FileScoresReversed),
    reverse(FileScoresReversed, FileScores).

% compare_histograms_helper(QueryHisto, Files, DatasetDirectory,
% AccFileScores, FileScores)
% computes histograms from the file list (with the dataset dir), then
% computes the 'relatedness' score, and adds that with the filename to a
% list
compare_histograms_helper(_, [], _, FileScores, FileScores).
compare_histograms_helper(QueryHisto, [File|Rest], DatasetDirectory, AccFileScores, FileScores) :-
    atom_concat(DatasetDirectory, File, FullPath),
    read_hist_file(FullPath, DatasetHisto),
    histogram_intersection(QueryHisto, DatasetHisto, Score),
    round_score(Score, RoundedScore),
    FileScore = file_score(File, RoundedScore),
    compare_histograms_helper(QueryHisto, Rest, DatasetDirectory, [FileScore|AccFileScores], FileScores).



% histogram_intersection(Histogram1, Histogram2, Score)
% compute the intersection similarity score between two histograms
% Score is between 0.0 and 1.0 (1.0 for identical histograms)
histogram_intersection([], [], 0).
histogram_intersection([Val1|QTail], [Val2|Tail], Score) :-
   Val is min(Val1, Val2),
   histogram_intersection(QTail, Tail, NewScore),
   Score is NewScore + Val.

% take(List,K,KList)
% extracts the K first items in a list
take(Src,N,L) :- findall(E, (nth1(I,Src,E), I =< N), L).

% atoms_numbers(ListOfAtoms,ListOfNumbers)
% converts a list of atoms into a list of numbers
atoms_numbers([],[]).
atoms_numbers([X|L],[Y|T]):- atom_number(X,Y), atoms_numbers(L,T).
atoms_numbers([X|L],T):- \+atom_number(X,_), atoms_numbers(L,T).


%=======================
%        HELPERS
%=======================


% normalize_histogram(Histogram, NormalizedHistogram)
% Normalizes a histogram by dividing each value by the sum of all values.
normalize_histogram(Histogram, NormalizedHistogram) :-
    sum_list(Histogram, Sum),
    maplist(divide_by_sum(Sum), Histogram, NormalizedHistogram).

% divide_by_sum(Sum, Value, NormalizedValue)
% Helper predicate to divide a value by the sum of all values in the histogram.
divide_by_sum(Sum, Value, NormalizedValue) :-
    NormalizedValue is Value / Sum.


% round_to_two_decimal_places(Number, RoundedNumber)
% Rounds a floating-point number to two decimal places.
round_score(Number, RoundedNumber) :-
    format(atom(RoundedAtom), '~2f', [Number]),
    atom_number(RoundedAtom, RoundedNumber).








