import numpy
import numpy
from numpy import ndarray
from sklearn.cluster import kmeans_plusplus 
from dataclasses import dataclass, asdict
import json
import gzip

@dataclass
class TransactionRaw:
    vector: ndarray
    label: str

@dataclass
class Transaction: 
    vector: list[float]
    legit: bool

def get_centers(transactions: list[TransactionRaw]) -> np.ndarray:    
    arrays = list()
    for t in transactions:
        arrays.append(t.vector)
    X = numpy.vstack(arrays)
    centers, idxs = kmeans_plusplus(X, n_clusters=20000)
    return centers

def load_to_json(vectors: list[numpy.ndarray], legit: bool, result: list[TransactionRaw]):
    for v in vectors:
        t = Transaction(v.tolist(), legit)
        result.append(asdict(t))
        print(asdict(t))
def main():
    with gzip.open('../resources/references.json.gz', 'rt', encoding='utf-8') as f:
        data = json.load(f)

    transactionsLegit = []
    transactionsFraud  = []
    for d in data:
        t = TransactionRaw(**d)
        if t.label == "legit": 
            transactionsLegit.append(t)
            continue
        transactionsFraud.append(t)



    centers_fraud = get_centers(transactionsFraud)
    centers_legit = get_centers(transactionsLegit)
    result = []
    load_to_json(centers_legit, True, result)
    load_to_json(centers_fraud, False, result)
    final_json = json.dumps(result)
    print(final_json)
    with open("../resources/references.json", "w", encoding='utf-8') as f:
        f.write(final_json)


main()



