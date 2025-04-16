// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract CRSManager {
    struct CRSInfo {
        uint256 round;
        bytes crs; // Store the CRS itself
        uint256 timestamp;
        address[] participants;
    }

    struct Commitment {
        address participant;
        bytes32 commitment;
        bool submitted;
    }

    uint256 public currentRound;
    uint256 public roundDuration; // Duration of each CRS round in seconds
    uint256 public maxParticipants;
    uint256 public commitDeadline;

    mapping(uint256 => CRSInfo) public crsHistory;
    mapping(uint256 => mapping(address => Commitment)) public commitments;
    mapping(uint256 => address[]) public registeredParticipants;

    bytes public currentCRS; 
    uint256 public currentContributorIdx; 
    bool public ceremonyActive; 

    event Registered(address indexed participant, uint256 round);
    event Committed(address indexed participant, uint256 round, bytes32 commitment);
    event Finalized(uint256 round, bytes crs, address[] participants);
    event CRSContributed(address indexed participant, uint256 round, bytes newCRS);

    constructor(uint256 _roundDuration, uint256 _maxParticipants) {
        roundDuration = _roundDuration;
        maxParticipants = _maxParticipants;
        currentRound = 1;
        commitDeadline = block.timestamp + roundDuration;
    }

    // Register for the next CRS round
    function register() external {
        require(block.timestamp < commitDeadline, "Registration closed");
        require(registeredParticipants[currentRound].length <= maxParticipants, "Max participants reached");
        address[] storage participants = registeredParticipants[currentRound];
        for (uint i = 0; i < participants.length; i++) {
            require(participants[i] != msg.sender, "Already registered");
        }
        participants.push(msg.sender);
        // If first participant, initialize ceremony state
        if (participants.length == 1) {
            currentContributorIdx = 0;
            delete currentCRS;
            ceremonyActive = true;
        }
        emit Registered(msg.sender, currentRound);
    }

    // Submit commitment for this round
    function submitCommitment(bytes32 commitment) external {
        require(block.timestamp < commitDeadline, "Commitment phase ended");
        address[] storage participants = registeredParticipants[currentRound];
        bool isParticipant = false;
        for (uint i = 0; i < participants.length; i++) {
            if (participants[i] == msg.sender) {
                isParticipant = true;
                break;
            }
        }
        require(isParticipant, "Not a registered participant");
        Commitment storage c = commitments[currentRound][msg.sender];
        require(!c.submitted, "Already submitted");
        c.participant = msg.sender;
        c.commitment = commitment;
        c.submitted = true;
        emit Committed(msg.sender, currentRound, commitment);
    }

    // Sequential CRS contribution
    function contributeCRS(bytes calldata newCRS) external {
        require(ceremonyActive, "No active ceremony");
        address[] storage participants = registeredParticipants[currentRound];
        require(participants.length > 0, "No participants");
        require(currentContributorIdx < participants.length, "All have contributed");
        require(participants[currentContributorIdx] == msg.sender, "Not your turn");
        // Accept the CRS
        currentCRS = newCRS;
        emit CRSContributed(msg.sender, currentRound, newCRS);
        currentContributorIdx++;
    }

    // Finalize CRS round and anchor CRS
    function finalizeCRS() external {
        require(ceremonyActive, "No active ceremony");
        address[] storage participants = registeredParticipants[currentRound];
        require(participants.length > 0, "No participants");
        require(currentContributorIdx == participants.length, "Not all have contributed");
        CRSInfo storage info = crsHistory[currentRound];
        require(info.crs.length == 0, "Already finalized");
        info.round = currentRound;
        info.crs = currentCRS;
        info.timestamp = block.timestamp;
        for (uint i = 0; i < participants.length; i++) {
            info.participants.push(participants[i]);
        }
        emit Finalized(currentRound, currentCRS, participants);
        // Prepare for next round
        currentRound += 1;
        commitDeadline = block.timestamp + roundDuration;
        ceremonyActive = false;
        delete currentCRS;
        currentContributorIdx = 0;
    }

    // Get latest CRS
    function getLatestCRS() external view returns (bytes memory, uint256, address[] memory) {
        CRSInfo storage info = crsHistory[currentRound - 1];
        return (info.crs, info.timestamp, info.participants);
    }

    // Get registered participants for current round
    function getRegisteredParticipants() external view returns (address[] memory) {
        return registeredParticipants[currentRound];
    }

    // Getter for currentContributorIdx
    function getCurrentContributorIdx() public view returns (uint256) {
        return currentContributorIdx;
    }

    // Getter for currentCRS (in-progress CRS value)
    function getCurrentCRS() public view returns (bytes memory) {
        return currentCRS;
    }
}
